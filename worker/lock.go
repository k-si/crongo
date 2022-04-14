package worker

import (
	"context"
	"github.com/k-si/crongo/common"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
)

type JobLock struct {
	lockPath        string
	leaseId         clientv3.LeaseID
	cancelKeepAlive context.CancelFunc
	isLocked        bool
}

func CreateJobLock(jobName string) *JobLock {
	return &JobLock{
		lockPath: common.LockDir + jobName,
	}
}

func (lock *JobLock) Lock() (err error) {
	var (
		grantResp         *clientv3.LeaseGrantResponse
		keepAliveRespChan <-chan *clientv3.LeaseKeepAliveResponse
		txnResp           *clientv3.TxnResponse
		txn               clientv3.Txn
	)

	// 自动续租
	if grantResp, err = EtcdConn.cli.Lease.Grant(context.TODO(), 5); err != nil {
		return
	}
	keepCtx, keepCancel := context.WithCancel(context.TODO())
	if keepAliveRespChan, err = EtcdConn.cli.KeepAlive(keepCtx, grantResp.ID); err != nil {
		goto Rollback
	}

	// 消费租约
	go func() {
		var (
			keepAliveResp *clientv3.LeaseKeepAliveResponse
		)
		for {
			select {
			case keepAliveResp = <-keepAliveRespChan:
				if keepAliveResp == nil {
					log.Println("[", lock.lockPath, "] stop to keep lease alive, leaseID:", grantResp.ID)
					goto end
				}
			}
		}
	end:
	}()

	// 事务(set if not exist) 实现加锁
	txn = EtcdConn.cli.Txn(context.TODO())
	txn.If(clientv3.Compare(clientv3.CreateRevision(lock.lockPath), "=", 0)).
		Then(clientv3.OpPut(lock.lockPath, "", clientv3.WithLease(grantResp.ID)))
	if txnResp, err = txn.Commit(); err != nil {
		goto Rollback
	}
	if !txnResp.Succeeded {
		err = common.ErrTxnLockFail
		goto Rollback
	}

	log.Println("[", lock.lockPath, "] lock success")
	lock.leaseId = grantResp.ID
	lock.cancelKeepAlive = keepCancel
	lock.isLocked = true

	return

	// 锁资源回收
Rollback:
	log.Println("[", lock.lockPath, "] lock rollback,", err)
	keepCancel()
	EtcdConn.cli.Revoke(context.TODO(), grantResp.ID) // 删除租约失败，租约过几秒也会自动撤销

	return
}

func (lock *JobLock) UnLock() {
	log.Println("[", lock.lockPath, "] unlock")
	if lock.isLocked {
		lock.cancelKeepAlive()
		EtcdConn.cli.Revoke(context.TODO(), lock.leaseId)
		lock.isLocked = false
	}
}
