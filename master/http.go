package master

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/k-si/crongo/common"
	"net/http"
	"strconv"
	"time"
)

var HttpServer http.Server

func InitHttpServer() {
	HttpServer = http.Server{
		Addr:         fmt.Sprintf(":%d", Cfg.Port),
		Handler:      NewRouter(),
		ReadTimeout:  time.Duration(Cfg.ReadTimeOut) * time.Millisecond,
		WriteTimeout: time.Duration(Cfg.WriteTimeOut) * time.Millisecond,
	}
}

func NewRouter() *gin.Engine {

	r := gin.Default()

	job := r.Group("/job")
	{
		job.POST("/save", JobSave)
		job.GET("/delete/:name", JobDelete)
		job.GET("/list", JobList)
		job.GET("/interrupt/:name", JobInterrupt)
	}
	jLog := r.Group("/log")
	{
		jLog.GET("/list/:name", LogList)
	}
	worker := r.Group("/worker")
	{
		worker.GET("/list", WorkerList)
	}

	return r
}

func JobSave(ctx *gin.Context) {
	var (
		job common.Job
		err error
	)
	if err = ctx.ShouldBindJSON(&job); err != nil {
		common.Response(ctx, common.CodeInvalidParam, nil, nil)
		return
	}
	if err = EtcdConn.SaveJob(&job); err != nil {
		common.Response(ctx, common.CodeInternalError, err.Error(), nil)
		return
	}
	common.Response(ctx, common.CodeSuccess, nil, nil)
}

func JobDelete(ctx *gin.Context) {
	var err error
	jobName := ctx.Param("name")
	if jobName == "" {
		common.Response(ctx, common.CodeInvalidParam, "缺少Job名称", nil)
		return
	}
	if err = EtcdConn.DeleteJob(jobName); err != nil {
		common.Response(ctx, common.CodeInternalError, err.Error(), nil)
		return
	}
	common.Response(ctx, common.CodeSuccess, nil, nil)
}

func JobList(ctx *gin.Context) {
	var (
		err  error
		jobs []*common.Job
	)
	if jobs, err = EtcdConn.ListJob(); err != nil {
		common.Response(ctx, common.CodeInternalError, err.Error(), nil)
		return
	}
	common.Response(ctx, common.CodeSuccess, nil, jobs)
}

func JobInterrupt(ctx *gin.Context) {
	var err error
	jobName := ctx.Param("name")
	if jobName == "" {
		common.Response(ctx, common.CodeInvalidParam, "缺少Job名称", nil)
		return
	}
	if err = EtcdConn.InterruptJob(jobName); err != nil {
		common.Response(ctx, common.CodeInternalError, err.Error(), nil)
		return
	}
	common.Response(ctx, common.CodeSuccess, nil, nil)
}

func LogList(ctx *gin.Context) {
	var (
		err    error
		iSkip  int
		iLimit int
		logs   []*common.JobLog
	)

	name := ctx.Param("name")
	skip := ctx.Query("skip")
	limit := ctx.Query("limit")

	// request param check
	if name == "" {
		common.Response(ctx, common.CodeInvalidParam, "缺少Job名称", nil)
		return
	}
	if skip == "" {
		skip = "0"
	}
	if limit == "" {
		limit = "20"
	}
	if iSkip, err = strconv.Atoi(skip); err != nil {
		common.Response(ctx, common.CodeInvalidParam, "skip参数不合法", nil)
		return
	}
	if iLimit, err = strconv.Atoi(limit); err != nil {
		common.Response(ctx, common.CodeInvalidParam, "limit参数不合法", nil)
		return
	}

	if logs, err = MongoConn.LogList(name, iSkip, iLimit); err != nil {
		common.Response(ctx, common.CodeInternalError, err.Error(), nil)
		return
	}
	common.Response(ctx, common.CodeSuccess, nil, logs)
}

// db.log.insert({"name":"test", "command":"echo -n hello", "output":"hello", "error_info":"", "start_time":1649860519, "end_time":1649860519, "expected_schedule_time":1649860519, "real_schedule_time":1649860519})

func WorkerList(ctx *gin.Context) {
	var (
		err     error
		workers []*common.Worker
	)

	if workers, err = EtcdConn.ListWorker(); err != nil {
		common.Response(ctx, common.CodeInternalError, err.Error(), nil)
		return
	}
	common.Response(ctx, common.CodeSuccess, nil, workers)
}
