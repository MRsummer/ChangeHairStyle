package model

import (
	"time"
)

// HairStyleRecord 发型生成记录
type HairStyleRecord struct {
	ID        int64     `json:"id"`
	UserID    string    `json:"user_id"`    // 用户ID
	ImageURL  string    `json:"image_url"`  // 生成的图片URL
	Prompt    string    `json:"prompt"`     // 使用的提示词
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// RecordResponse 记录列表响应
type RecordResponse struct {
	Total   int64             `json:"total"`   // 总记录数
	Records []HairStyleRecord `json:"records"` // 记录列表
} 