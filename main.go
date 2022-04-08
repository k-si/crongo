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

	// 初始化scheduler
	ctx, cancel := context.WithCancel(context.Background())
	server.InitJobScheduler(ctx)
	defer cancel()

	// 启动worker监听
	ctx2, cancel2 := context.WithCancel(context.Background())
	go server.WatchJob(ctx2)
	defer cancel2()

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
	ctx3, cancel3 := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel3()
	if err = server.HttpServer.Shutdown(ctx3); err != nil {
		log.Println("http shutdown error:", err)
		return
	}

	log.Println("server stopped!")
}
