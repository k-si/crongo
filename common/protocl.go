package common

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Job struct {
	Name    string `json:"name"`
	Command string `json:"command"`
	Express string `json:"express"`
}

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
