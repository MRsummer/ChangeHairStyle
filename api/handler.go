package handler

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/MRsummer/ChangeHairStyle/pkg/handler"
	"github.com/MRsummer/ChangeHairStyle/pkg/volcengine"
)

var r *gin.Engine

func init() {
	// 从环境变量获取密钥
	accessKeyID := os.Getenv("VOLCENGINE_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("VOLCENGINE_SECRET_ACCESS_KEY")

	if accessKeyID == "" || secretAccessKey == "" {
		log.Fatal("请设置环境变量 VOLCENGINE_ACCESS_KEY_ID 和 VOLCENGINE_SECRET_ACCESS_KEY")
	}

	// 创建火山引擎客户端
	client := volcengine.NewClient(accessKeyID, secretAccessKey)

	// 创建处理器
	hairStyleHandler := handler.NewHairStyleHandler(client)

	// 创建Gin引擎
	r = gin.Default()

	// 设置路由
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// 发型生成路由
	r.POST("/api/hair-style", hairStyleHandler.Generate)
	r.POST("/api/hair-style/base64", hairStyleHandler.GenerateWithBase64)
}

// Handler 是 Vercel Serverless 函数的入口点
func Handler(w http.ResponseWriter, req *http.Request) {
	r.ServeHTTP(w, req)
} 