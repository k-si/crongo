package worker

import (
	"context"
	"log"
	"os/exec"
)

var Executor JobExecutor

type JobExecutor struct {
	PlanChan chan *JobPlan
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
	var (
		plan   *JobPlan
		cmd    *exec.Cmd
		output []byte
		err    error
	)

	for {
		select {
		case <-ctx.Done():
			goto END
		default:
			for plan = range e.PlanChan {

				if plan.Status == Waiting {
					// todo:抢分布式锁

					// 如果抢锁成功
					plan.Status = Running

					// 执行job
					cmd = exec.Command("/bin/bash", "-c", plan.Job.Command)
					output, err = cmd.CombinedOutput()
					if err != nil {
						log.Println("[", plan.Job.Name, "]", "err:", err.Error())
					}
					if output != nil {
						log.Println("[", plan.Job.Name, "]", "output:", string(output))
					}

					// 释放锁
					plan.Status = Waiting
				}
			}
		}
	}
END:
}
