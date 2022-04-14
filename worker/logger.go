package worker

import (
	"context"
	"github.com/k-si/crongo/common"
	"log"
	"time"
)

var Logger JobLogger

type JobLogger struct {
	jobLogChan chan *common.JobLog
}

func LogJob(ctx context.Context) {
	Logger = JobLogger{
		jobLogChan: make(chan *common.JobLog, Cfg.JobLogChanSize),
	}
	go Logger.Log(ctx)
}

func (l JobLogger) Log(ctx context.Context) {

	var (
		jl     *common.JobLog
		timer  = time.NewTimer(time.Duration(Cfg.JobLogSendInterval) * time.Millisecond)
		bundle = make([]interface{}, 0)
	)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			goto end
		default:

			// 两个阀值：n秒发送一次或n条发送一次
			for jl = range l.jobLogChan {
				bundle = append(bundle, jl)
				select {
				case <-timer.C:
					_ = l.WriteLog(bundle)
					timer.Reset(time.Duration(Cfg.JobLogSendInterval) * time.Millisecond)
					bundle = bundle[0:0]
				default:
					if len(bundle) >= Cfg.JobLogBundleSize {
						_ = l.WriteLog(bundle)
						timer.Reset(time.Duration(Cfg.JobLogSendInterval) * time.Millisecond)
						bundle = bundle[0:0]
					}
				}
			}
		}
	}
end:
}

func (l JobLogger) PushJobLog(jl *common.JobLog) {
	l.jobLogChan <- jl
}

func (l JobLogger) WriteLog(bundle []interface{}) (err error) {
	if _, err = MongoConn.collection.InsertMany(context.TODO(), bundle); err != nil {
		log.Println("bundle log write fail, err:", err)
	}
	return
}
