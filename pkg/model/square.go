package model

import "time"

// SquareContent 广场内容
type SquareContent struct {
	ID        int64     `json:"id"`
	UserID    string    `json:"user_id"`
	RecordID  int64     `json:"record_id"`
	LikeCount int       `json:"like_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联的发型记录信息
	Record *HairStyleRecord `json:"record,omitempty"`
	// 关联的用户信息
	UserInfo *UserInfo `json:"user_info,omitempty"`
	// 当前用户是否已点赞
	IsLiked bool `json:"is_liked"`
}

// SquareContentResponse 广场内容列表响应
type SquareContentResponse struct {
	Total   int64           `json:"total"`   // 总记录数
	Records []SquareContent `json:"records"` // 记录列表
}

// ShareToSquareRequest 分享到广场请求
type ShareToSquareRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	RecordID int64  `json:"record_id" binding:"required"`
}

// LikeContentRequest 点赞请求
type LikeContentRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	ContentID int64  `json:"content_id" binding:"required"`
}
