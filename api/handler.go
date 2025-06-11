package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/MRsummer/ChangeHairStyle/pkg/handler"
)

var r *gin.Engine

func init() {
	// 设置 Gin 为发布模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎
	r = gin.New()
	r.Use(gin.Recovery())

	// 设置路由
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// 发型生成路由
	r.POST("/api/hair-style", handler.HandleHairStyle)
}

// main 函数是程序入口点
func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		r.ServeHTTP(w, req)
	})
	log.Fatal(http.ListenAndServe(":9000", nil))
}

// Handler 是云函数入口点
func Handler(ctx context.Context, req []byte) ([]byte, error) {
	// 解析请求
	var request struct {
		Path       string            `json:"path"`
		Method     string            `json:"method"`
		Headers    map[string]string `json:"headers"`
		Query      map[string]string `json:"query"`
		Body       string            `json:"body"`
		IsBase64   bool             `json:"isBase64"`
	}

	if err := json.Unmarshal(req, &request); err != nil {
		return nil, err
	}

	// 创建响应
	response := struct {
		StatusCode int               `json:"statusCode"`
		Headers    map[string]string `json:"headers"`
		Body       string            `json:"body"`
		IsBase64   bool             `json:"isBase64"`
	}{
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	// 处理请求
	w := &responseWriter{
		headers: make(map[string]string),
		body:    make([]byte, 0),
	}

	// 创建请求上下文
	c := &gin.Context{
		Request: &http.Request{
			Method: request.Method,
			URL: &url.URL{
				Path: request.Path,
			},
			Header: http.Header{},
		},
	}

	// 设置请求头
	for k, v := range request.Headers {
		c.Request.Header.Set(k, v)
	}

	// 设置请求体
	if request.Body != "" {
		c.Request.Body = ioutil.NopCloser(strings.NewReader(request.Body))
	}

	// 处理请求
	r.ServeHTTP(w, c.Request)

	// 设置响应
	response.StatusCode = w.statusCode
	response.Body = string(w.body)
	for k, v := range w.headers {
		response.Headers[k] = v
	}

	// 返回响应
	return json.Marshal(response)
}

// responseWriter 实现 http.ResponseWriter 接口
type responseWriter struct {
	statusCode int
	headers    map[string]string
	body       []byte
}

func (w *responseWriter) Header() http.Header {
	return http.Header{}
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return len(data), nil
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
} 