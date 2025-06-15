package model

import "time"

// UserInfo 用户信息
type UserInfo struct {
	ID             int64      `json:"id"`
	UserID         string     `json:"user_id"`
	Nickname       string     `json:"nickname"`
	AvatarURL      string     `json:"avatar_url"`
	Coin           int        `json:"coin"`
	InviteCode     string     `json:"invite_code"`
	UsedInviteCode string     `json:"used_invite_code"`
	LastSignInDate *time.Time `json:"last_sign_in_date,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// UpdateUserInfoRequest 更新用户信息请求
type UpdateUserInfoRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

// GenerateInviteCodeRequest 生成邀请码请求
type GenerateInviteCodeRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// UseInviteCodeRequest 使用邀请码请求
type UseInviteCodeRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	InviteCode string `json:"invite_code" binding:"required"`
}

// SignInRequest 签到请求
type SignInRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// WxLoginRequest 微信登录请求
type WxLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

// WxLoginResponse 微信登录响应
type WxLoginResponse struct {
	UserID    string `json:"user_id"`
	Nickname  string `json:"nickname,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Coin      int    `json:"coin"`
	Token     string `json:"token"`
}
