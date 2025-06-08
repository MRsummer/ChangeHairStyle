package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"wx-backend/internal/handler"
	"wx-backend/pkg/volcengine"
)

func main() {
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
	r := gin.Default()

	// 设置路由
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// 发型生成路由
	r.POST("/api/hair-style", hairStyleHandler.Generate)
	r.POST("/api/hair-style/base64", hairStyleHandler.GenerateWithBase64)

	// 获取端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 启动服务器
	if err := r.Run(":" + port); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
} 