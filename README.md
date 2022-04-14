# Introduction

The project is a distributed task scheduling service, which is mainly used to schedule and execute Linux crontab tasks.
The service can start one master node and multiple worker nodes, Store crontab tasks through etcd and provide
distributed locks for multiple workers. The master module is responsible for the addition, deletion, modification and
query of tasks on the HTTP interface, and the worker node is responsible Monitor etcd changes and maintain a scheduling
table in memory. Multiple worker nodes can execute tasks concurrently.

External exposure interface:

```
Save job interface:
POST /job/save
{
    "name":"test",   
    "command":"echo hello",
    "express":"*/5 * * * * * *"
}

Delete job interface:
GET /job/delete/{name}

Query job interface:
GET /job/list

interrupt job interface:
GET /job/interrupt/{name}

Query job log interface:
GET /log/list/{name}?skip={0}&limit={20}
```