package worker

import (
	"context"
	"fmt"
	"log"
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
	Err       error
	StartTime time.Time
	EndTime   time.Time
}

func ExecuteJob(ctx context.Context) {
	Executor = JobExecutor{
		PlanChan: make(chan *JobPlan, 1000),
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
		lock = CreateJobLock(plan.Job.Name)
		if err = lock.Lock(); err != nil {
			return
		}
		plan.Status = Running

		// 执行job
		start = time.Now()
		cmd = exec.Command("/bin/bash", "-c", plan.Job.Command)
		output, err = cmd.CombinedOutput()
		end = time.Now()

		// 释放锁
		lock.UnLock()
		plan.Status = Waiting

		res = &JobResult{
			Plan:      plan,
			Output:    output,
			Err:       err,
			StartTime: start,
			EndTime:   end,
		}

		// todo: 执行结果存储
		info := fmt.Sprintf("[ %s ] finished \n "+
			"output: %s \n "+
			"expected schedule time: %s \n "+
			"real schedule time: %s \n "+
			"job start time: %s \n "+
			"job end time: %s",
			res.Plan.Job.Name,
			res.Output,
			res.Plan.Expected.Format(time.RFC3339),
			res.Plan.Real.Format(time.RFC3339),
			res.StartTime.Format(time.RFC3339),
			res.EndTime.Format(time.RFC3339))
		log.Println(info)

	} else {
		log.Println("skip [", plan.Job.Name, "] it's still running...")
	}
}
