package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/MRsummer/ChangeHairStyle/pkg/db"
	"github.com/MRsummer/ChangeHairStyle/pkg/handler"
	"github.com/MRsummer/ChangeHairStyle/pkg/logger"
	"github.com/MRsummer/ChangeHairStyle/pkg/middleware"
	"github.com/gin-gonic/gin"
)

var r *gin.Engine
var database *sql.DB

func init() {
	// 初始化日志系统
	logger.Init()

	// 初始化数据库连接
	var err error
	database, err = db.InitDB()
	if err != nil {
		logger.Fatalf("初始化数据库失败: %v", err)
	}

	// 设置 Gin 为发布模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎
	r = gin.New()
	r.Use(gin.Recovery())

	// 添加请求追踪中间件
	r.Use(middleware.TraceMiddleware())

	// 添加数据库中间件
	r.Use(func(c *gin.Context) {
		c.Set("db", database)
		c.Next()
	})

	// 设置路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
		})
	})

	// 发型生成路由
	r.POST("/api/hair-style", handler.HandleHairStyle)

	// 获取生成记录路由
	r.GET("/api/hair-style/records", handler.HandleGetRecords)

	// 用户信息路由
	r.POST("/api/user/info", handler.HandleUpdateUserInfo)
	r.GET("/api/user/info/get", handler.HandleGetUserInfo)
	r.POST("/api/user/code/use", handler.HandleUseInviteCode)
	r.POST("/api/user/sign-in", handler.HandleSignIn)
	r.POST("/api/user/wx-login", handler.HandleWxLogin)

	// 广场相关路由
	r.POST("/api/square/share", handler.HandleShareToSquare)
	r.GET("/api/square/contents", handler.HandleGetSquareContents)
	r.POST("/api/square/like", handler.HandleLike)
}

// main 函数是程序入口点
func main() {
	defer database.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		r.ServeHTTP(w, req)
	})
	logger.Info("服务器启动，监听端口 9000")
	log.Fatal(http.ListenAndServe(":9000", nil))
}

// Handler 是云函数入口点
func Handler(ctx context.Context, req []byte) ([]byte, error) {
	// 解析请求
	var request struct {
		Path     string            `json:"path"`
		Method   string            `json:"method"`
		Headers  map[string]string `json:"headers"`
		Query    map[string]string `json:"query"`
		Body     string            `json:"body"`
		IsBase64 bool              `json:"isBase64"`
	}

	if err := json.Unmarshal(req, &request); err != nil {
		return nil, err
	}

	// 创建响应
	response := struct {
		StatusCode int               `json:"statusCode"`
		Headers    map[string]string `json:"headers"`
		Body       string            `json:"body"`
		IsBase64   bool              `json:"isBase64"`
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
	httpReq := &http.Request{
		Method: request.Method,
		URL: &url.URL{
			Path:     request.Path,
			RawQuery: buildQueryString(request.Query),
		},
		Header: http.Header{},
	}

	// 设置请求头
	for k, v := range request.Headers {
		httpReq.Header.Set(k, v)
	}

	// 设置请求体
	if request.Body != "" {
		httpReq.Body = ioutil.NopCloser(strings.NewReader(request.Body))
	}

	// 处理请求
	r.ServeHTTP(w, httpReq)

	// 设置响应
	response.StatusCode = w.statusCode
	response.Body = string(w.body)
	for k, v := range w.headers {
		response.Headers[k] = v
	}

	// 返回响应
	return json.Marshal(response)
}

// buildQueryString 构建查询字符串
func buildQueryString(queryParams map[string]string) string {
	if len(queryParams) == 0 {
		return ""
	}

	values := url.Values{}
	for k, v := range queryParams {
		values.Add(k, v)
	}

	return values.Encode()
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
