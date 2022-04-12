package worker

import (
	"context"
	"encoding/json"
	"github.com/k-si/crongo/common"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"strings"
)

const (
	// job event
	Save   = 0
	Delete = 1
)

var Watcher JobWatcher

type JobWatcher struct {
}

type JobEvent struct {
	Opt int
	Job *common.Job
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

	// 将所有job调入内存
	if getResp, err = Connector.cli.Get(context.TODO(), common.JobDir, clientv3.WithPrefix()); err != nil {
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

	watchChan := Connector.cli.Watch(context.TODO(), common.JobDir, clientv3.WithRev(getResp.Header.Revision+1), clientv3.WithPrefix())

	// 监听job变动
	for {
		select {
		case <-ctx.Done():
			goto end
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
end:

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
				Name: strings.TrimPrefix(string(e.Kv.Key), common.JobDir),
			},
		}
	}

	Scheduler.PushJobEvent(je)
}
