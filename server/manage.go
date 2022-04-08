package server

import (
	"context"
	"encoding/json"
	"github.com/k-si/crongo/common"
	"github.com/k-si/crongo/conf"
	"go.etcd.io/etcd/client/v3"
	"time"
)

const (
	JobDir  = "/cron/job/"
	KillDir = "/cron/kill/"
)

var Manager JobManager

type JobManager struct {
	cfg clientv3.Config
	cli *clientv3.Client
}

func InitJobManager() (err error) {
	cfg := clientv3.Config{
		Endpoints:   conf.Cfg.Endpoints,
		DialTimeout: time.Duration(conf.Cfg.DialTimeOut) * time.Millisecond,
	}
	cli, err := clientv3.New(cfg)
	Manager.cfg = cfg
	Manager.cli = cli
	return
}

func (mgr JobManager) End() error {
	return mgr.cli.Close()
}

func (mgr JobManager) SaveJob(job *common.Job) (err error) {

	var b []byte

	if b, err = json.Marshal(job); err != nil {
		return
	}
	if _, err = mgr.cli.Put(context.TODO(), JobDir+job.Name, string(b)); err != nil {
		return
	}

	return
}

func (mgr JobManager) DeleteJob(jobName string) (err error) {
	if _, err = mgr.cli.Delete(context.TODO(), JobDir+jobName); err != nil {
		return
	}
	return
}

func (mgr JobManager) ListJob() (jobs []*common.Job, err error) {

	var getResp *clientv3.GetResponse
	jobs = make([]*common.Job, 0)

	if getResp, err = mgr.cli.Get(context.TODO(), JobDir, clientv3.WithPrefix()); err != nil {
		return
	}

	for _, kv := range getResp.Kvs {
		job := &common.Job{}
		if err = json.Unmarshal(kv.Value, job); err != nil {
			return
		}
		jobs = append(jobs, job)
	}

	return
}

func (mgr JobManager) KillJob(jobName string) (err error) {

	var (
		grantResp *clientv3.LeaseGrantResponse
	)

	if grantResp, err = mgr.cli.Grant(context.TODO(), 1); err != nil {
		return
	}
	if _, err = mgr.cli.Put(context.TODO(), KillDir+jobName, "", clientv3.WithLease(grantResp.ID)); err != nil {
		return
	}

	return
}