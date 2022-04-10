package main

import (
	"context"
	"flag"
	"github.com/k-si/crongo/conf"
	"github.com/k-si/crongo/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	configPath string
)

func InitArgs() {
	flag.StringVar(&configPath, "c", "./config.yaml", "config file path")
	flag.Parse()
}

func main() {
	var err error

	// 初始化命令行参数
	InitArgs()

	// 初始化配置文件
	if err = conf.InitConfig(configPath); err != nil {
		log.Println("config init error:", err)
		return
	}

	// 初始化manager
	if err = server.InitJobManager(); err != nil {
		log.Println("job manager init error:", err)
		return
	}
	defer server.Manager.End()

	// 开启job调度
	scheduleCtx, scheduleCancel := context.WithCancel(context.Background())
	defer scheduleCancel()
	server.ScheduleJob(scheduleCtx)

	// 开启job执行
	executeCtx, executeCancel := context.WithCancel(context.Background())
	defer executeCancel()
	server.ExecuteJob(executeCtx)

	// 开启job变动监听
	watchCtx, watchCancel := context.WithCancel(context.Background())
	defer watchCancel()
	server.WatchJob(watchCtx)

	// 启动http服务
	go func() {
		server.InitHttpServer()
		if err = server.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("http serve error:", err)
			return
		}
	}()

	// 监听信号
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	// 停止服务
	shutdownCtx, shutdownCancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer shutdownCancel()
	if err = server.HttpServer.Shutdown(shutdownCtx); err != nil {
		log.Println("http shutdown error:", err)
		return
	}

	log.Println("server stopped!")
}
