package server

import (
	"context"
	"encoding/json"
	"github.com/k-si/crongo/common"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"strings"
)

var Watcher JobWatcher

type JobWatcher struct {
}

func WatchJob(ctx context.Context) {
	Watcher = JobWatcher{}
	go Watcher.Watch(ctx)
}

func (w JobWatcher) Watch(ctx context.Context) {

	var (
		err     error
		getResp *clientv3.GetResponse
	)

	if getResp, err = Manager.cli.Get(context.TODO(), JobDir, clientv3.WithPrefix()); err != nil {
		log.Println("get revision err:", err)
		return
	}
	for _, kv := range getResp.Kvs {
		job := &common.Job{}
		if err = json.Unmarshal(kv.Value, job); err != nil {
			log.Println("job Unmarshal fail:", err)
			continue
		}
		Scheduler.PushJobEvent(&JobEvent{Save, job})
	}

	watchChan := Manager.cli.Watch(context.TODO(), JobDir, clientv3.WithRev(getResp.Header.Revision+1), clientv3.WithPrefix())

	for {
		select {
		case <-ctx.Done():
			goto END
		default:
			for watchResp := range watchChan {
				for _, event := range watchResp.Events {
					switch event.Type {
					case mvccpb.PUT:
						pushJobEvent(Save, event) // 交由调度模块处理
					case mvccpb.DELETE:
						pushJobEvent(Delete, event)
					}
				}
			}
		}
	}
END:

	return
}

func pushJobEvent(opt int, e *clientv3.Event) {
	var je *JobEvent

	if opt == Save {
		job := &common.Job{}
		if err := json.Unmarshal(e.Kv.Value, job); err != nil {
			log.Println("pushJobEvent Unmarshal fail:", opt, err)
			return
		}
		je = &JobEvent{opt, job}
	} else if opt == Delete {
		je = &JobEvent{
			opt,
			&common.Job{
				Name: strings.TrimPrefix(string(e.Kv.Key), JobDir),
			},
		}
	}

	Scheduler.PushJobEvent(je)
}
