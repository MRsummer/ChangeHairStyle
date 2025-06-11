package cos

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"net/url"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// Client 腾讯云 COS 客户端
type Client struct {
	client *cos.Client
	bucket string
	region string
}

// NewClient 创建腾讯云 COS 客户端
func NewClient(secretID, secretKey, bucket, region string) (*Client, error) {
	// 创建 COS 客户端配置
	u := fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket, region)
	b := &cos.BaseURL{BucketURL: mustParseURL(u)}
	
	// 创建 COS 客户端
	c := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
		},
	})

	return &Client{
		client: c,
		bucket: bucket,
		region: region,
	}, nil
}

func mustParseURL(u string) *url.URL {
	parsed, err := url.Parse(u)
	if err != nil {
		panic(err)
	}
	return parsed
}

// FetchImage 从URL获取图片并上传到腾讯云 COS
func (c *Client) FetchImage(imageURL string) (string, error) {
	// 1. 从URL获取图片
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("获取图片失败: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取图片失败，状态码: %d", resp.StatusCode)
	}

	// 2. 生成唯一的文件名
	fileName := fmt.Sprintf("hair_style/%d.jpg", time.Now().UnixNano())

	// 3. 上传到腾讯云 COS
	_, err = c.client.Object.Put(context.Background(), fileName, resp.Body, nil)
	if err != nil {
		return "", fmt.Errorf("上传到腾讯云 COS 失败: %v", err)
	}

	// 4. 返回可访问的URL（使用 HTTPS）
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", c.bucket, c.region, fileName), nil
} 