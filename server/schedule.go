package server

import (
	"context"
	"github.com/gorhill/cronexpr"
	"github.com/k-si/crongo/common"
	"log"
	"time"
)

const (
	Save   = 0
	Delete = 1
)

var Scheduler JobScheduler

type JobEvent struct {
	Opt int
	Job *common.Job
}

type JobPlan struct {
	Job  *common.Job
	Expr *cronexpr.Expression
	Next time.Time // 下次执行时间
}

type JobScheduler struct {
	JobEventChan chan *JobEvent
	JobPlanTable map[string]*JobPlan
}

func InitJobScheduler(ctx context.Context) {
	Scheduler = JobScheduler{
		JobEventChan: make(chan *JobEvent, 1000),
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
		Job:  job,
		Expr: expr,
		Next: expr.Next(time.Now()),
	}
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
			goto END
		}
		interval = sdr.TryScheduling()
		t.Reset(interval)
	}
END:
}

func (sdr JobScheduler) HandleJobEvent(je *JobEvent) {
	switch je.Opt {
	case Save:
		jp, err := NewJobPlan(je.Job)
		if err != nil {
			return
		}
		sdr.JobPlanTable[je.Job.Name] = jp
	case Delete:
		delete(sdr.JobPlanTable, je.Job.Name)
	}
}

// 每次只需等待最短的过期时间
// 例如三个任务分别为1s、2s、3s执行一次，当前时刻为0s
// now  a    b    c
// 0s - 1s - 2s - 3s
// 此时只需等待1s
func (sdr JobScheduler) TryScheduling() (interval time.Duration) {

	now := time.Now()
	var near *time.Time

	for name, plan := range sdr.JobPlanTable {

		if plan.Next.Before(now) || plan.Next.Equal(now) {
			log.Println(name)
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
