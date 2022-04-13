package worker

import (
	"context"
	"github.com/gorhill/cronexpr"
	"github.com/k-si/crongo/common"
	"log"
	"time"
)

const (
	// job status
	Waiting = 0
	Running = 1
)

var Scheduler JobScheduler

type JobPlan struct {
	Job        *common.Job
	Expr       *cronexpr.Expression
	Next       time.Time // 下次调度时间
	Expected   time.Time // 期望调度时间
	Real       time.Time // 实际调度时间
	Status     int
	CancelCtx  context.Context // command命令及时被kill
	CancelFunc context.CancelFunc
}

type JobScheduler struct {
	JobEventChan chan *JobEvent
	JobPlanTable map[string]*JobPlan
}

func ScheduleJob(ctx context.Context) {
	Scheduler = JobScheduler{
		JobEventChan: make(chan *JobEvent, Cfg.JobEventChanSize),
		JobPlanTable: make(map[string]*JobPlan),
	}
	go Scheduler.Plan(ctx)
}

func NewJobPlan(job *common.Job) (jp *JobPlan, err error) {
	var expr *cronexpr.Expression
	if expr, err = cronexpr.Parse(job.Express); err != nil {
		log.Println("parse expression error", err)
		return
	}
	jp = &JobPlan{
		Job:    job,
		Expr:   expr,
		Next:   expr.Next(time.Now()),
		Status: Waiting,
	}
	jp.CancelCtx, jp.CancelFunc = context.WithCancel(context.TODO())
	return
}

func (sdr JobScheduler) PushJobEvent(je *JobEvent) {
	sdr.JobEventChan <- je
}

func (sdr JobScheduler) Plan(ctx context.Context) {

	var je *JobEvent

	interval := sdr.TryScheduling()
	t := time.NewTimer(interval)
	defer t.Stop()

	for {
		select {
		case je = <-sdr.JobEventChan:
			sdr.HandleJobEvent(je) // 修正调度表
		case <-t.C:
		case <-ctx.Done():
			goto end
		}
		interval = sdr.TryScheduling()
		t.Reset(interval)
	}
end:
}

func (sdr JobScheduler) HandleJobEvent(je *JobEvent) {
	switch je.Opt {
	case common.SaveJob:
		jp, err := NewJobPlan(je.Job)
		if err != nil {
			log.Println("save [ ", je.Job.Name, " ] fail, err:", err)
			return
		}
		sdr.JobPlanTable[je.Job.Name] = jp
	case common.DeleteJob:
		delete(sdr.JobPlanTable, je.Job.Name)
	case common.KillJob:
		jp, ok := sdr.JobPlanTable[je.Job.Name]
		if ok {
			if jp.Status == Running {
				log.Println("[", je.Job.Name, "] killed during running")
				jp.CancelFunc()
				// todo: 关于kill的功能需要再细化分析
				//jp.CancelCtx, jp.CancelFunc = context.WithCancel(context.TODO())
			}
		} else {
			log.Println("try kill [", je.Job.Name, "] , but not in jobPlanTable")
		}
	}
}

// 每次只需等待最短的过期时间
// 例如三个任务分别为1s、2s、3s执行一次，当前时刻为0s
// now  a    b    c
// 0s - 1s - 2s - 3s
// 此时只需等待1s
func (sdr JobScheduler) TryScheduling() (interval time.Duration) {

	var (
		name string
		plan *JobPlan
	)

	now := time.Now()
	var near *time.Time

	for name, plan = range sdr.JobPlanTable {

		// 执行过期任务
		if plan.Next.Before(now) || plan.Next.Equal(now) {
			log.Println("[", name, "]", "scheduled success")
			plan.Expected = plan.Next
			plan.Real = now
			Executor.PushJobPlan(plan) // 交由executor执行任务
			plan.Next = plan.Expr.Next(now)
		}

		// 找到最短过期时间
		if near == nil || plan.Next.Before(*near) {
			near = &plan.Next
		}
	}

	// plan table为空时，固定休眠间隔为1s
	if near == nil {
		interval = time.Second
		return
	}

	interval = (*near).Sub(time.Now())
	return
}
