package worker

import (
	"context"
	"github.com/k-si/crongo/common"
)

var Logger JobLogger

type JobLogger struct {
	jobLogChan chan *common.JobLog
}

func LogJob(ctx context.Context) {
	Logger = JobLogger{}
	go Logger.Log(ctx)
}

func (l JobLogger) Log(ctx context.Context) {

	var (
		jl *common.JobLog
	)

	for {
		select {
		case <-ctx.Done():
			goto end
		default:
			for jl = range l.jobLogChan {
				// todo: mongodb batch save
				jl = jl
			}
		}
	}
end:
}

func (l JobLogger) PushJobResult(res *common.JobLog) {
	l.jobLogChan <- res
}
