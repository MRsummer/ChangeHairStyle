package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// User 用户结构体
type User struct {
	ID      int
	UserID  string
	BaseURL string
	LogChan chan string
	Stats   *TestStats
}

// LoginResponse 登录响应
type LoginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		UserID string `json:"user_id"`
	} `json:"data"`
}

// TestStats 测试统计
type TestStats struct {
	TotalUsers      int64
	SuccessUsers    int64
	FailedUsers     int64
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	StartTime       time.Time
	EndTime         time.Time
	mu              sync.Mutex
}

// 增加成功用户数
func (ts *TestStats) IncSuccessUsers() {
	atomic.AddInt64(&ts.SuccessUsers, 1)
}

// 增加失败用户数
func (ts *TestStats) IncFailedUsers() {
	atomic.AddInt64(&ts.FailedUsers, 1)
}

// 增加成功请求数
func (ts *TestStats) IncSuccessRequests() {
	atomic.AddInt64(&ts.SuccessRequests, 1)
}

// 增加失败请求数
func (ts *TestStats) IncFailedRequests() {
	atomic.AddInt64(&ts.FailedRequests, 1)
}

// 增加总请求数
func (ts *TestStats) IncTotalRequests() {
	atomic.AddInt64(&ts.TotalRequests, 1)
}

// 模拟用户操作
func (u *User) simulate() {
	u.log("开始测试...")
	success := false
	defer func() {
		if success {
			u.Stats.IncSuccessUsers()
		} else {
			u.Stats.IncFailedUsers()
		}
	}()

	// 1. 微信登录
	u.log("步骤1: 微信登录")
	loginData := map[string]string{
		"code": u.UserID,
	}
	loginResp, err := u.post("/api/user/wx-login", loginData)
	if err != nil {
		u.log(fmt.Sprintf("登录失败: %v", err))
		return
	}

	var loginResponse LoginResponse
	if err := json.Unmarshal(loginResp, &loginResponse); err != nil {
		u.log(fmt.Sprintf("解析登录响应失败: %v", err))
		return
	}

	if loginResponse.Code != 0 {
		u.log(fmt.Sprintf("登录失败: %s", loginResponse.Message))
		return
	}

	userID := loginResponse.Data.UserID
	u.log(fmt.Sprintf("登录成功，获取到user_id: %s", userID))

	// 等待1秒
	time.Sleep(1 * time.Second)

	// 2. 换发型
	u.log("步骤2: 换发型")
	hairStyleData := map[string]interface{}{
		"user_id":   userID,
		"image_url": fmt.Sprintf("https://example.com/test_%d.jpg", u.ID),
		"prompt":    "短发",
	}
	u.post("/api/hair-style", hairStyleData)
	u.log("换发型完成")

	time.Sleep(1 * time.Second)

	// 3. 获取记录
	u.log("步骤3: 获取记录")
	u.get(fmt.Sprintf("/api/hair-style/records?user_id=%s", userID))
	u.log("获取记录完成")

	time.Sleep(1 * time.Second)

	// 4. 获取用户信息
	u.log("步骤4: 获取用户信息")
	u.get(fmt.Sprintf("/api/user/info/get?user_id=%s", userID))
	u.log("获取用户信息完成")

	time.Sleep(1 * time.Second)

	// 5. 更新用户信息
	u.log("步骤5: 更新用户信息")
	updateData := map[string]interface{}{
		"user_id":    userID,
		"nickname":   fmt.Sprintf("测试用户%d", u.ID),
		"avatar_url": fmt.Sprintf("https://example.com/avatar_%d.jpg", u.ID),
	}
	u.post("/api/user/info", updateData)
	u.log("更新用户信息完成")

	time.Sleep(1 * time.Second)

	// 6. 再次获取用户信息
	u.log("步骤6: 再次获取用户信息")
	u.get(fmt.Sprintf("/api/user/info/get?user_id=%s", userID))
	u.log("再次获取用户信息完成")

	time.Sleep(1 * time.Second)

	// 7. 签到
	u.log("步骤7: 签到")
	signInData := map[string]string{
		"user_id": userID,
	}
	u.post("/api/user/sign-in", signInData)
	u.log("签到完成")

	time.Sleep(1 * time.Second)

	// 8. 分享到广场
	u.log("步骤8: 分享到广场")
	shareData := map[string]interface{}{
		"user_id":   userID,
		"record_id": 150,
	}
	u.post("/api/square/share", shareData)
	u.log("分享完成")

	time.Sleep(1 * time.Second)

	// 9. 获取广场内容
	u.log("步骤9: 获取广场内容")
	u.get(fmt.Sprintf("/api/square/contents?cursor=0&page_size=10&user_id=%s", userID))
	u.log("获取广场内容完成")

	time.Sleep(1 * time.Second)

	// 10. 点赞
	u.log("步骤10: 点赞")
	likeData := map[string]interface{}{
		"user_id":    userID,
		"content_id": 1,
	}
	u.post("/api/square/like", likeData)
	u.log("点赞完成")

	u.log("测试完成")
	success = true
}

// 发送POST请求
func (u *User) post(path string, data interface{}) ([]byte, error) {
	u.Stats.IncTotalRequests()

	jsonData, err := json.Marshal(data)
	if err != nil {
		u.Stats.IncFailedRequests()
		return nil, err
	}

	resp, err := http.Post(u.BaseURL+path, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		u.Stats.IncFailedRequests()
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		u.Stats.IncFailedRequests()
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		u.Stats.IncSuccessRequests()
	} else {
		u.Stats.IncFailedRequests()
	}

	u.log(fmt.Sprintf("POST %s - 状态码: %d", path, resp.StatusCode))
	return body, nil
}

// 发送GET请求
func (u *User) get(path string) ([]byte, error) {
	u.Stats.IncTotalRequests()

	resp, err := http.Get(u.BaseURL + path)
	if err != nil {
		u.Stats.IncFailedRequests()
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		u.Stats.IncFailedRequests()
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		u.Stats.IncSuccessRequests()
	} else {
		u.Stats.IncFailedRequests()
	}

	u.log(fmt.Sprintf("GET %s - 状态码: %d", path, resp.StatusCode))
	return body, nil
}

// 记录日志
func (u *User) log(message string) {
	logMsg := fmt.Sprintf("[用户%d] %s", u.ID, message)
	u.LogChan <- logMsg
}

// 打印统计信息
func (ts *TestStats) PrintStats() {
	duration := ts.EndTime.Sub(ts.StartTime)

	fmt.Println("\n=== 测试统计信息 ===")
	fmt.Printf("总用户数: %d\n", ts.TotalUsers)
	fmt.Printf("成功用户数: %d\n", ts.SuccessUsers)
	fmt.Printf("失败用户数: %d\n", ts.FailedUsers)
	fmt.Printf("成功率: %.2f%%\n", float64(ts.SuccessUsers)/float64(ts.TotalUsers)*100)

	fmt.Printf("\n总请求数: %d\n", ts.TotalRequests)
	fmt.Printf("成功请求数: %d\n", ts.SuccessRequests)
	fmt.Printf("失败请求数: %d\n", ts.FailedRequests)
	fmt.Printf("请求成功率: %.2f%%\n", float64(ts.SuccessRequests)/float64(ts.TotalRequests)*100)

	fmt.Printf("\n总耗时: %v\n", duration)
	fmt.Printf("平均每个用户耗时: %v\n", duration/time.Duration(ts.TotalUsers))
	if ts.TotalRequests > 0 {
		fmt.Printf("平均每个请求耗时: %v\n", duration/time.Duration(ts.TotalRequests))
	}
}

func main() {
	var baseURL string
	var userCount int
	var verbose bool

	flag.StringVar(&baseURL, "url", "https://1255379329-gl8iz72lbx.ap-guangzhou.tencentscf.com", "服务器基础URL")
	flag.IntVar(&userCount, "users", 100, "用户数量")
	flag.BoolVar(&verbose, "verbose", true, "详细输出模式")
	flag.Parse()

	stats := &TestStats{
		TotalUsers: int64(userCount),
		StartTime:  time.Now(),
	}

	fmt.Printf("=== 用户场景测试开始 ===\n")
	fmt.Printf("基础URL: %s\n", baseURL)
	fmt.Printf("用户数量: %d\n", userCount)
	fmt.Printf("详细模式: %v\n", verbose)
	fmt.Printf("开始时间: %s\n", stats.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Println()

	// 创建日志通道
	logChan := make(chan string, userCount*100)
	var wg sync.WaitGroup

	// 启动日志打印协程
	go func() {
		for logMsg := range logChan {
			if verbose {
				fmt.Println(logMsg)
			}
		}
	}()

	// 创建并启动所有用户
	for i := 1; i <= userCount; i++ {
		user := &User{
			ID:      i,
			UserID:  fmt.Sprintf("test_user_%d_%d", i, time.Now().Unix()),
			BaseURL: baseURL,
			LogChan: logChan,
			Stats:   stats,
		}

		wg.Add(1)
		go func(u *User) {
			defer wg.Done()
			u.simulate()
		}(user)
	}

	// 等待所有用户完成
	wg.Wait()

	// 关闭日志通道
	close(logChan)

	// 记录结束时间
	stats.EndTime = time.Now()

	fmt.Println()
	fmt.Printf("=== 所有用户测试完成 ===\n")
	fmt.Printf("完成时间: %s\n", stats.EndTime.Format("2006-01-02 15:04:05"))

	// 打印统计信息
	stats.PrintStats()
}
