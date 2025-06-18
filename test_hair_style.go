package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// User 用户结构体
type User struct {
	ID        int
	UserID    string
	BaseURL   string
	LogChan   chan string
	Stats     *TestStats
	ImageData string
	Prompt    string
}

// HairStyleResponse 发型生成响应
type HairStyleResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		RecordID int    `json:"record_id"`
		ImageURL string `json:"image_url"`
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
}

// 预定义的发型提示词
var hairPrompts = []string{
	"短发",
	"长发",
	"卷发",
	"直发",
	"波浪发",
	"蓬松发",
	"紧贴头皮",
	"高马尾",
	"低马尾",
	"丸子头",
}

// 预定义的用户ID
var userIDs = []string{
	"test_user_15_1750268703",
	"test_user_90_1750268703",
	"test_user_84_1750268703",
	"test_user_17_1750268703",
	"test_user_86_1750268703",
	"test_user_78_1750268703",
	"test_user_81_1750268703",
	"test_user_48_1750268703",
	"test_user_57_1750268703",
	"test_user_2_1750268703",
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

// 读取图片文件并转换为base64
func readImageAsBase64(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 读取文件内容
	imageData, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	// 转换为base64
	base64Data := base64.StdEncoding.EncodeToString(imageData)
	return base64Data, nil
}

// 模拟用户操作
func (u *User) simulate() {
	u.log(fmt.Sprintf("开始测试，使用发型: %s", u.Prompt))
	success := false
	defer func() {
		if success {
			u.Stats.IncSuccessUsers()
		} else {
			u.Stats.IncFailedUsers()
		}
	}()

	// 发送发型生成请求
	u.log("发送发型生成请求")
	hairStyleData := map[string]interface{}{
		"user_id":      u.UserID,
		"base64_image": u.ImageData,
		"prompt":       u.Prompt,
	}

	resp, err := u.post("/api/hair-style", hairStyleData)
	if err != nil {
		u.log(fmt.Sprintf("请求失败: %v", err))
		return
	}

	var hairStyleResponse HairStyleResponse
	if err := json.Unmarshal(resp, &hairStyleResponse); err != nil {
		u.log(fmt.Sprintf("解析响应失败: %v", err))
		return
	}

	if hairStyleResponse.Code != 0 {
		u.log(fmt.Sprintf("发型生成失败: %s", hairStyleResponse.Message))
		return
	}

	u.log(fmt.Sprintf("发型生成成功，记录ID: %d", hairStyleResponse.Data.RecordID))
	if hairStyleResponse.Data.ImageURL != "" {
		u.log(fmt.Sprintf("原始图片链接: %s", hairStyleResponse.Data.ImageURL))
	}
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

// 记录日志
func (u *User) log(message string) {
	logMsg := fmt.Sprintf("[用户%d-%s] %s", u.ID, u.UserID, message)
	u.LogChan <- logMsg
}

// 打印统计信息
func (ts *TestStats) PrintStats() {
	duration := ts.EndTime.Sub(ts.StartTime)

	fmt.Println("\n=== 发型生成测试统计信息 ===")
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
	var imagePath string
	var verbose bool

	flag.StringVar(&baseURL, "url", "https://1255379329-gl8iz72lbx.ap-guangzhou.tencentscf.com", "服务器基础URL")
	flag.StringVar(&imagePath, "image", "~/Downloads/pic/test.jpeg", "图片文件路径")
	flag.BoolVar(&verbose, "verbose", true, "详细输出模式")
	flag.Parse()

	// 展开用户主目录
	if imagePath[:2] == "~/" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("获取用户主目录失败: %v\n", err)
			return
		}
		imagePath = homeDir + imagePath[1:]
	}

	// 读取图片文件
	fmt.Printf("正在读取图片文件: %s\n", imagePath)
	imageData, err := readImageAsBase64(imagePath)
	if err != nil {
		fmt.Printf("读取图片文件失败: %v\n", err)
		return
	}
	fmt.Printf("图片读取成功，大小: %d bytes\n", len(imageData))

	stats := &TestStats{
		TotalUsers: int64(len(userIDs)),
		StartTime:  time.Now(),
	}

	fmt.Printf("\n=== 发型生成接口测试开始 ===\n")
	fmt.Printf("基础URL: %s\n", baseURL)
	fmt.Printf("用户数量: %d\n", len(userIDs))
	fmt.Printf("详细模式: %v\n", verbose)
	fmt.Printf("开始时间: %s\n", stats.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Println()

	// 创建日志通道
	logChan := make(chan string, len(userIDs)*10)
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
	for i, userID := range userIDs {
		user := &User{
			ID:        i + 1,
			UserID:    userID,
			BaseURL:   baseURL,
			LogChan:   logChan,
			Stats:     stats,
			ImageData: imageData,
			Prompt:    hairPrompts[i], // 每个用户使用不同的发型提示词
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
