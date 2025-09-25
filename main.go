package main

import (
	"backend/control"
	"backend/log"
	"backend/svc"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

func main() {

	config := svc.Config{
		SystemLog:  log.Log("systemLog"),  //程序日志
		RequestLog: log.Log("requestLog"), //请求日志
		Request:    nil,                   //构建的请求
		C:          nil,                   //config
	}

	s := gin.Default()

	s.Static("/assets", "./static/dist/assets")
	s.StaticFile("/index", "./static/dist/index.html")

	s.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:9999"}, // 前端地址
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := s.Group("api")
	{
		api.GET("/request", control.AllControl(config))
		api.POST("/request", control.AllControl(config))
		api.DELETE("/request", control.AllControl(config))
		api.PUT("/request", control.AllControl(config))
		//api.PATCH("/request", func(c *gin.Context) {
		//
		//})
	}

	err := s.Run(":7744")
	if err != nil {
		config.SystemLog.Fatal(err) //输出日志并退出程序
	}

}
