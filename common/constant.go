package common

const (

	// etcd dir
	JobDir       = "/cron/job/"
	InterruptDir = "/cron/kill/"
	LockDir      = "/cron/lock/"

	// job event
	SaveJob      = 0
	DeleteJob    = 1
	InterruptJob = 2
)
