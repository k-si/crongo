package common

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// vo object
type Job struct {
	Name    string `json:"name"`
	Command string `json:"command"`
	Express string `json:"express"`
}

type JobLog struct {
	Name                 string `json:"name" bson:"name"`
	Command              string `json:"command" bson:"command"`
	Output               string `json:"output" bson:"output"`
	ErrorInfo            string `json:"error_info" bson:"error_info"`
	StartTime            int64  `json:"start_time" bson:"start_time"`
	EndTime              int64  `json:"end_time" bson:"end_time"`
	ExpectedScheduleTime int64  `json:"expected_schedule_time" bson:"expected_schedule_time"`
	RealScheduleTime     int64  `json:"real_schedule_time" bson:"real_schedule_time"`
}

type Worker struct {
	IP string `json:"ip"`
}

// http response format object
type RespCode int64

const (
	CodeSuccess RespCode = 1000 + iota
	CodeInvalidParam
	CodeInternalError
)

var CodeMsgMap = map[RespCode]string{
	CodeSuccess:       "request success",
	CodeInvalidParam:  "invalid param",
	CodeInternalError: "internal error",
}

type R struct {
	Code RespCode    `json:"code"`
	Msg  interface{} `json:"msg"`
	Data interface{} `json:"data"`
}

func (rc RespCode) mapping() (msg interface{}) {
	var ok bool
	msg, ok = CodeMsgMap[rc]
	if !ok {
		msg = CodeMsgMap[CodeInternalError]
	}
	return
}

func Response(ctx *gin.Context, code RespCode, msg interface{}, data interface{}) {
	if msg == nil {
		msg = code.mapping()
	}
	ctx.JSON(http.StatusOK, R{
		Code: code,
		Msg:  msg,
		Data: data,
	})
}
