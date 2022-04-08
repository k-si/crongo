package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/k-si/crongo/common"
	"github.com/k-si/crongo/conf"
	"net/http"
	"time"
)

var HttpServer http.Server

func InitHttpServer() {
	HttpServer = http.Server{
		Addr:         fmt.Sprintf(":%d", conf.Cfg.Port),
		Handler:      NewRouter(),
		ReadTimeout:  time.Duration(conf.Cfg.ReadTimeOut) * time.Millisecond,
		WriteTimeout: time.Duration(conf.Cfg.WriteTimeOut) * time.Millisecond,
	}
}

func NewRouter() *gin.Engine {

	r := gin.Default()

	job := r.Group("/job")
	{
		job.POST("/save", JobSave)
		job.GET("/delete", JobDelete)
		job.GET("/list", JobList)
		job.GET("/kill", JobKill)
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
	}
	if err = Manager.SaveJob(&job); err != nil {
		common.Response(ctx, common.CodeInternalError, err.Error(), nil)
	}
	common.Response(ctx, common.CodeSuccess, nil, nil)
}

func JobDelete(ctx *gin.Context) {
	var err error
	jobName := ctx.Query("name")
	if jobName == "" {
		common.Response(ctx, common.CodeInvalidParam, nil, nil)
	}
	if err = Manager.DeleteJob(jobName); err != nil {
		common.Response(ctx, common.CodeInternalError, err.Error(), nil)
	}
	common.Response(ctx, common.CodeSuccess, nil, nil)
}

func JobList(ctx *gin.Context) {
	var (
		err  error
		jobs []*common.Job
	)
	if jobs, err = Manager.ListJob(); err != nil {
		common.Response(ctx, common.CodeInternalError, err.Error(), nil)
	}
	common.Response(ctx, common.CodeSuccess, nil, jobs)
}

func JobKill(ctx *gin.Context) {
	var err error
	jobName := ctx.Query("name")
	if jobName == "" {
		common.Response(ctx, common.CodeInvalidParam, nil, nil)
	}
	if err = Manager.KillJob(jobName); err != nil {
		common.Response(ctx, common.CodeInternalError, err.Error(), nil)
	}
	common.Response(ctx, common.CodeSuccess, nil, nil)
}
