package worker

import (
	"context"
	"github.com/k-si/crongo/common"
	"log"
	"math/rand"
	"os/exec"
	"time"
)

var Executor JobExecutor

type JobExecutor struct {
	PlanChan chan *JobPlan
}

type JobResult struct {
	Plan      *JobPlan
	Output    []byte
	Err       string
	StartTime time.Time
	EndTime   time.Time
}

func ExecuteJob(ctx context.Context) {
	Executor = JobExecutor{
		PlanChan: make(chan *JobPlan, Cfg.JobPlanChanSize),
	}
	go Executor.Execute(ctx)
}

func (e JobExecutor) PushJobPlan(jp *JobPlan) {
	e.PlanChan <- jp
}

func (e JobExecutor) Execute(ctx context.Context) {
	var plan *JobPlan

	for {
		select {
		case <-ctx.Done():
			goto END
		default:
			for plan = range e.PlanChan {
				go e.RunPlan(plan)
			}
		}
	}
END:
}

func (e JobExecutor) RunPlan(plan *JobPlan) {
	var (
		output []byte
		err    error
		jl     *common.JobLog
	)

	if plan.Status == Waiting {

		// 不同机器时钟校验有微秒级的差异，为了任务的均衡调度，在允许任务调度有轻微延迟时，随机睡眠0-1000ms
		// 当任务中最小执行时间小于睡眠时间，会导致不同节点的重复执行
		if Cfg.BalanceOptimization {
			time.Sleep(time.Duration(rand.Intn(Cfg.BalanceSleepTime)) * time.Millisecond)
		}

		// 抢分布式锁
		lock := CreateJobLock(plan.Job.Name)
		if err = lock.Lock(); err != nil {
			return
		}
		plan.Status = Running

		// 执行job
		start := time.Now()
		cmd := exec.CommandContext(plan.CancelCtx, "/bin/bash", "-c", plan.Job.Command)
		output, err = cmd.CombinedOutput()
		end := time.Now()

		// 释放锁
		lock.UnLock()
		plan.Status = Waiting

		// 交由logger存储结果日志
		jl = &common.JobLog{
			Name:                 plan.Job.Name,
			Command:              plan.Job.Command,
			Output:               string(output),
			StartTime:            start.Unix(),
			EndTime:              end.Unix(),
			ExpectedScheduleTime: plan.Expected.Unix(),
			RealScheduleTime:     plan.Real.Unix(),
		}
		if err == nil {
			jl.ErrorInfo = ""
		} else {
			jl.ErrorInfo = err.Error()
		}
		Logger.PushJobLog(jl)

	} else {
		log.Println("[", plan.Job.Name, "] skipped, it's still running...")
	}
}
