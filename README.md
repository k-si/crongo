# crongo

crontab job scheduling service

```
Save job interface:
POST /job/save
{
    "name":"test",   
    "command":"echo hello",
    "express":"*/5 * * * * * *"
}

Delete job interface:
GET /job/delete? name=test

Query job interface:
GET /job/list

Kill job interface:
GET /job/kill? name=test
```