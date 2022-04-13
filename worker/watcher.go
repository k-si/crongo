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

var Watcher JobWatcher

type JobWatcher struct {
}

type JobEvent struct {
	Opt int
	Job *common.Job
}

func WatchJob(ctx context.Context) {
	Watcher = JobWatcher{}
	go Watcher.WatchJobDir(ctx)
	go Watcher.watchKillDir(ctx)
}

func (w JobWatcher) WatchJobDir(ctx context.Context) {
	var (
		err     error
		getResp *clientv3.GetResponse
	)

	// 将所有job调入内存
	if getResp, err = EtcdConn.cli.Get(context.TODO(), common.JobDir, clientv3.WithPrefix()); err != nil {
		log.Println("watch jobDir, get revision err:", err)
		return
	}
	for _, kv := range getResp.Kvs {
		job := &common.Job{}
		if err = json.Unmarshal(kv.Value, job); err != nil {
			log.Println("job Unmarshal fail:", err)
			continue
		}
		Scheduler.PushJobEvent(&JobEvent{common.SaveJob, job})
	}

	watchChan := EtcdConn.cli.Watch(context.TODO(), common.JobDir, clientv3.WithRev(getResp.Header.Revision+1), clientv3.WithPrefix())

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
						pushJobEvent(common.SaveJob, event) // 交由调度模块处理
					case mvccpb.DELETE:
						pushJobEvent(common.DeleteJob, event)
					}
				}
			}
		}
	}
end:

	return
}

func (w JobWatcher) watchKillDir(ctx context.Context) {

	// 由于killDir所有键都自动租约过期，可认为启动worker时，killDir为空，无需指定版本号监听
	watchChan := EtcdConn.cli.Watch(context.TODO(), common.KillDir, clientv3.WithPrefix())

	for {
		select {
		case <-ctx.Done():
			goto end
		default:
			for watchResp := range watchChan {
				for _, event := range watchResp.Events {
					switch event.Type {
					case mvccpb.PUT:
						pushJobEvent(common.KillJob, event)
					case mvccpb.DELETE:
						// kill目录下的key有租约可自动过期，无需处理delete事件
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

	if opt == common.SaveJob {
		job := &common.Job{}
		if err := json.Unmarshal(e.Kv.Value, job); err != nil {
			log.Println("pushJobEvent Unmarshal fail:", opt, err)
			return
		}
		je = &JobEvent{opt, job}
	} else if opt == common.DeleteJob {
		je = &JobEvent{
			opt,
			&common.Job{
				Name: strings.TrimPrefix(string(e.Kv.Key), common.JobDir),
			},
		}
	} else if opt == common.KillJob {
		je = &JobEvent{
			opt,
			&common.Job{
				Name: strings.TrimPrefix(string(e.Kv.Key), common.KillDir),
			},
		}
	}

	Scheduler.PushJobEvent(je)
}
