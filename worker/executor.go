package worker

import (
	"context"
	"fmt"
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
		cmd    *exec.Cmd
		start  time.Time
		end    time.Time
		output []byte
		err    error
		lock   *JobLock
		res    *JobResult
	)

	if plan.Status == Waiting {

		// 不同机器时钟校验有微秒级的差异，为了任务的均衡调度，在允许任务调度有轻微延迟时，随机睡眠0-1000ms
		// 当任务执行很快时，可能会导致短时间内不同节点的重复执行
		if Cfg.BalanceOptimization {
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		}

		lock = CreateJobLock(plan.Job.Name)
		if err = lock.Lock(); err != nil {
			return
		}
		plan.Status = Running

		// 执行job
		start = time.Now()
		cmd = exec.CommandContext(plan.CancelCtx, "/bin/bash", "-c", plan.Job.Command)
		output, err = cmd.CombinedOutput()
		end = time.Now()

		// 释放锁
		lock.UnLock()
		plan.Status = Waiting

		res = &JobResult{
			Plan:      plan,
			Output:    output,
			StartTime: start,
			EndTime:   end,
		}
		if err == nil {
			res.Err = ""
		} else {
			res.Err = err.Error()
		}

		// todo: 执行结果存储
		info := fmt.Sprintf("[ %s ] finished \n "+
			"output: %s \n "+
			"error: %s \n"+
			"expected schedule time: %s \n "+
			"real schedule time: %s \n "+
			"job start time: %s \n "+
			"job end time: %s",
			res.Plan.Job.Name,
			res.Output,
			res.Err,
			res.Plan.Expected.Format(time.RFC3339),
			res.Plan.Real.Format(time.RFC3339),
			res.StartTime.Format(time.RFC3339),
			res.EndTime.Format(time.RFC3339))
		log.Println(info)

	} else {
		log.Println("[", plan.Job.Name, "] skipped, it's still running...")
	}
}
