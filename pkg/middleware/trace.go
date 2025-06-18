package middleware

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MRsummer/ChangeHairStyle/pkg/logger"
	"github.com/gin-gonic/gin"
)

// responseWriter 包装响应写入器以捕获响应内容
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// TraceMiddleware 请求追踪中间件
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 生成请求ID
		requestID := generateRequestID()
		c.Set("request_id", requestID)

		// 设置请求ID到响应头
		c.Header("X-Request-ID", requestID)

		// 创建响应写入器包装器
		responseWriter := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = responseWriter

		// 创建日志上下文
		logCtx := map[string]interface{}{
			"request_id": requestID,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"user_agent": c.Request.UserAgent(),
			"client_ip":  c.ClientIP(),
		}

		logger.WithContext(logCtx).Info("请求开始")

		// 处理请求
		c.Next()

		// 记录请求结束
		duration := time.Since(startTime)
		status := c.Writer.Status()

		logCtx["duration_ms"] = duration.Milliseconds()
		logCtx["status_code"] = status

		// 根据状态码选择日志级别
		if status >= 500 {
			// 5xx错误：记录详细信息
			if len(c.Errors) > 0 {
				logCtx["errors"] = c.Errors.String()
			}
			// 尝试解析响应中的错误信息
			if responseBody := responseWriter.body.String(); responseBody != "" {
				var response map[string]interface{}
				if err := json.Unmarshal([]byte(responseBody), &response); err == nil {
					if message, exists := response["message"]; exists {
						logCtx["response_error"] = message
					}
				}
			}
			logger.WithContext(logCtx).Error("请求处理失败")
		} else if status >= 400 {
			// 4xx错误：记录详细信息
			if len(c.Errors) > 0 {
				logCtx["errors"] = c.Errors.String()
			}
			// 尝试解析响应中的错误信息
			if responseBody := responseWriter.body.String(); responseBody != "" {
				var response map[string]interface{}
				if err := json.Unmarshal([]byte(responseBody), &response); err == nil {
					if message, exists := response["message"]; exists {
						logCtx["response_error"] = message
					}
				}
			}
			logger.WithContext(logCtx).Warn("请求处理异常")
		} else if duration > 10*time.Second {
			// 超时警告：10秒
			logger.WithContext(logCtx).Error("请求处理时间过长")
		} else {
			logger.WithContext(logCtx).Info("请求处理完成")
		}
	}
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// GetRequestID 从上下文中获取请求ID
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return ""
}
