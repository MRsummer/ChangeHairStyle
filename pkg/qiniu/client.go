package qiniu

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

// Client 七牛云客户端
type Client struct {
	mac        *qbox.Mac
	bucket     string
	domain     string
	uploader   *storage.FormUploader
	manager    *storage.BucketManager
	putPolicy  storage.PutPolicy
}

// NewClient 创建七牛云客户端
func NewClient(accessKey, secretKey, bucket, domain string) *Client {
	mac := qbox.NewMac(accessKey, secretKey)
	cfg := storage.Config{
		Region: &storage.ZoneHuadong, // 根据你的存储区域选择
		UseHTTPS: true,
	}
	
	uploader := storage.NewFormUploader(&cfg)
	manager := storage.NewBucketManager(mac, &cfg)
	
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	putPolicy.Expires = 7200 // 2小时有效期
	
	return &Client{
		mac:       mac,
		bucket:    bucket,
		domain:    domain,
		uploader:  uploader,
		manager:   manager,
		putPolicy: putPolicy,
	}
}

// FetchImage 从URL获取图片并上传到七牛云
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
	
	// 3. 上传到七牛云
	ret := storage.PutRet{}
	err = c.uploader.Put(context.Background(), &ret, c.putPolicy.UploadToken(c.mac), fileName, resp.Body, resp.ContentLength, nil)
	if err != nil {
		return "", fmt.Errorf("上传到七牛云失败: %v", err)
	}
	
	// 4. 返回可访问的URL
	return fmt.Sprintf("http://%s/%s", c.domain, ret.Key), nil
} 