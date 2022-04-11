# 简介

该项目是分布式任务调度服务，主要用来做linux crontab任务的调度执行。服务可以启动1个master节点和多个worker节点，
通过etcd存储crontab任务，并为多个worker提供分布式锁。master模块负责http接口对任务的增删改查，worker节点负责 监听etcd变化，并在内存维护一份调度表，多个worker节点可并发执行任务。

对外暴露接口：

```
保存任务接口：
POST /job/save
{
    "name":"test",
    "command":"echo hello",
    "express":"*/5 * * * * * *"
}

删除任务接口：
GET /job/delete?name=test

查询任务接口：
GET /job/list

杀死任务接口：
GET /job/kill?name=test
```
