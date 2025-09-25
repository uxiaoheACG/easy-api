package svc

import (
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type RequestData struct {
	Url       string            `json:"url"`
	Method    string            `json:"method"`
	Header    map[string]string `json:"header"`
	Body      interface{}       `json:"body"`      // 兼容 json/xml/text
	Form      map[string]string `json:"form"`      // application/x-www-form-urlencoded
	Files     map[string]string `json:"files"`     // 文件上传：key=字段名, value=路径
	Params    map[string]string `json:"params"`    // GET 参数
	Frequency int               `json:"frequency"` //频率
	TimeOut   int               `json:"timeout"`   //超时时间 单位s
}

type Config struct {
	SystemLog  *log.Logger
	RequestLog *log.Logger
	Request    *http.Request
	C          *gin.Context
	ApiContext *context.Context
}
