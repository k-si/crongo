package server

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

				if plan.Status == Running {
					continue
				}

				log.Println("[", plan.Job.Name, "]", "executed")

				plan.Status = Running

				cmd = exec.Command("/bin/bash", "-c", plan.Job.Command)
				output, err = cmd.CombinedOutput()
				if err != nil {
					log.Println("err:", err.Error())
				}
				if output != nil {
					log.Println("output:", string(output))
				}

				plan.Status = Waiting
			}
		}
	}
END:
}
