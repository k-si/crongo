# crongo

任务调度服务

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
