package common

const (

	// etcd dir
	JobDir       = "/cron/job/"
	InterruptDir = "/cron/kill/"
	LockDir      = "/cron/lock/"
	WorkerDir    = "/cron/worker"

	// job event
	SaveJob      = 0
	DeleteJob    = 1
	InterruptJob = 2
)
