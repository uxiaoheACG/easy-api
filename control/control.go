package control

import (
	"backend/registerNewRequest"
	"backend/svc"
	"github.com/gin-gonic/gin"
	"net/http"
)

// RequestData 这个结构体是-接收前端请求的数据，根据该数据来构建一个完整的请求的结构体

// AllControl 集中控制中心
func AllControl(config svc.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		//把数据绑定好，方便后续处理数据。如果绑定失败，就返回给前端然后终止此次前端请求的执行
		ReqDates := svc.RequestData{}
		err := c.ShouldBindJSON(&ReqDates)
		if err != nil {
			config.RequestLog.Println("绑定参数失败", err.Error())
			c.JSON(500, gin.H{
				"code": 500,
				"msg":  err.Error(),
			})
			return
		}

		//这里需要进行获取url
		url := ReqDates.Url
		//这里要扩展一个函数：构建出完整的url
		req, err := registerNewRequest.RegisterNewRequest(&ReqDates, config)
		if err != nil {
			config.RequestLog.Println("构建请求失败", err.Error())
			c.JSON(500, gin.H{
				"code": 500,
				"msg":  err.Error(),
			})
			return
		}
		config.Request = req
		config.C = c
		//打印日志，调用函数（形参是构建出的url）
		switch c.Request.Method {
		case http.MethodGet:
			config.SystemLog.Println("get请求:", url)
			Send(config, ReqDates)
		case http.MethodPost:
			config.SystemLog.Println("post请求:", url)
			Send(config, ReqDates)

		case http.MethodPut:
			config.SystemLog.Println("put请求:", url)
			Send(config, ReqDates)

		case http.MethodDelete:
			config.SystemLog.Println("delete请求:", url)
			Send(config, ReqDates)
		}
	}
}
