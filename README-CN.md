# 简介

该项目是分布式任务调度服务，主要用来做linux crontab任务的调度执行。通过etcd存储crontab任务，并为多个worker提供分布式锁。master模块负责http接口对任务的增删改查，worker节点负责监听etcd变化，并在内存维护一份调度表，多个worker节点可并发执行任务。

# 启动

服务依赖于etcd和mongodb，请先配置好环境。

启动master

```shell
cd master/main
go run main.go
```

启动worker

```shell
cd worker/main
go run worker.go
```

# Api

```
保存任务接口：
POST /job/save
{
    "name":"test",
    "command":"echo hello",
    "express":"*/5 * * * * * *"
}

删除任务接口：
GET /job/delete/{name}

查询任务接口：
GET /job/list

中断任务接口：
GET /job/interrupt/{name}

查询任务日志接口：
GET /log/list/{name}?skip={0}&limit={20}

查询worker节点接口：
GET /worker/list
```
