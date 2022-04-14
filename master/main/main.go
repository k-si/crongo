package main

import (
	"context"
	"flag"
	"github.com/k-si/crongo/master"
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
	defer log.Println("server stopped!")

	var err error

	// 初始化配置
	InitArgs()
	if err = master.InitConfig(configPath); err != nil {
		log.Println("config init error:", err)
		return
	}
	if err = master.InitEtcdConnector(); err != nil {
		log.Println("etcd init error:", err)
		return
	}
	defer master.EtcdConn.Close()
	if err = master.InitMongoConnector(); err != nil {
		log.Println("mongo init error:", err)
		return
	}

	// 启动http服务
	go func() {
		master.InitHttpServer()
		if err = master.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
	if err = master.HttpServer.Shutdown(shutdownCtx); err != nil {
		log.Println("http shutdown error:", err)
		return
	}
}
