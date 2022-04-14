package main

import (
	"context"
	"flag"
	"github.com/k-si/crongo/worker"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	configPath string
)

func InitArgs() {
	flag.StringVar(&configPath, "c", "./config.yaml", "config file path")
	flag.Parse()
}

func main() {
	defer log.Println("worker stopped!")

	var err error

	// 初始化配置参数
	InitArgs()
	if err = worker.InitConfig(configPath); err != nil {
		log.Println("config init error:", err)
		return
	}
	if err = worker.InitEtcdConnector(); err != nil {
		log.Println("etcd init error:", err)
		return
	}
	defer worker.EtcdConn.Close()
	if err = worker.InitMongoConnector(); err != nil {
		log.Println("mongodb init error:", err)
		return
	}
	defer worker.MongoConn.Close()

	// 开启job日志存储
	logCtx, logCancel := context.WithCancel(context.TODO())
	defer logCancel()
	worker.LogJob(logCtx)

	// 开启job调度
	scheduleCtx, scheduleCancel := context.WithCancel(context.Background())
	defer scheduleCancel()
	worker.ScheduleJob(scheduleCtx)

	// 开启job执行
	executeCtx, executeCancel := context.WithCancel(context.Background())
	defer executeCancel()
	worker.ExecuteJob(executeCtx)

	// 开启job变动监听
	watchCtx, watchCancel := context.WithCancel(context.Background())
	defer watchCancel()
	worker.WatchJob(watchCtx)

	// 阻塞等待退出
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
