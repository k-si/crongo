package common

const (

	// etcd dir
	JobDir  = "/cron/job/"
	KillDir = "/cron/kill/"
	LockDir = "/cron/lock/"

	// job event
	SaveJob   = 0
	DeleteJob = 1
	KillJob   = 2
)
