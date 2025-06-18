package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/MRsummer/ChangeHairStyle/pkg/db"
	"github.com/MRsummer/ChangeHairStyle/pkg/model"
	"github.com/gin-gonic/gin"
)

// HandleUpdateUserInfo 处理更新用户信息请求
func HandleUpdateUserInfo(c *gin.Context) {
	var req model.UpdateUserInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 更新用户信息
	userInfo := &model.UserInfo{
		UserID:    req.UserID,
		Nickname:  req.Nickname,
		AvatarURL: req.AvatarURL,
	}

	dbConn := c.MustGet("db").(*sql.DB)
	if err := db.UpdateUserInfo(dbConn, userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// HandleUseInviteCode 处理使用邀请码请求
func HandleUseInviteCode(c *gin.Context) {
	var req model.UseInviteCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	dbConn := c.MustGet("db").(*sql.DB)
	err := db.UseInviteCode(dbConn, req.UserID, req.InviteCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// HandleSignIn 处理签到请求
func HandleSignIn(c *gin.Context) {
	var req model.SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	dbConn := c.MustGet("db").(*sql.DB)
	err := db.SignIn(dbConn, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// WxLoginRequest 微信登录请求
type WxLoginRequest struct {
	Code      string `json:"code" binding:"required"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

// HandleWxLogin 处理微信登录请求
func HandleWxLogin(c *gin.Context) {
	var req model.WxLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 调用微信登录接口获取openid
	appID := os.Getenv("WX_APP_ID")
	appSecret := os.Getenv("WX_APP_SECRET")
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appID, appSecret, req.Code)

	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": fmt.Sprintf("调用微信登录接口失败: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	var wxResp struct {
		OpenID     string `json:"openid"`
		SessionKey string `json:"session_key"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&wxResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": fmt.Sprintf("解析微信登录响应失败: %v", err),
		})
		return
	}

	if wxResp.ErrCode != 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": fmt.Sprintf("微信登录失败: %s", wxResp.ErrMsg),
		})
		return
	}

	// 生成用户ID（使用openid作为用户ID）
	userID := wxResp.OpenID

	// 查询用户信息
	dbConn := c.MustGet("db").(*sql.DB)
	userInfo, err := db.GetUserInfo(dbConn, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": fmt.Sprintf("获取用户信息失败: %v", err),
		})
		return
	}

	// 如果用户不存在，创建新用户
	if userInfo == nil {
		userInfo = &model.UserInfo{
			UserID: userID,
			Coin:   60, // 初始金币
		}
		if err := db.CreateUser(dbConn, userInfo); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": fmt.Sprintf("创建用户失败: %v", err),
			})
			return
		}
	}

	// 返回登录响应
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": model.WxLoginResponse{
			UserID:         userInfo.UserID,
			Nickname:       userInfo.Nickname,
			AvatarURL:      userInfo.AvatarURL,
			Coin:           userInfo.Coin,
			InviteCode:     userInfo.InviteCode,
			UsedInviteCode: userInfo.UsedInviteCode,
			LastSignInDate: userInfo.LastSignInDate,
		},
	})
}

// HandleGetUserInfo 处理获取用户信息请求
func HandleGetUserInfo(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误：user_id不能为空",
		})
		return
	}

	dbConn := c.MustGet("db").(*sql.DB)
	userInfo, err := db.GetUserInfo(dbConn, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": fmt.Sprintf("获取用户信息失败: %v", err),
		})
		return
	}

	if userInfo == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": model.GetUserInfoResponse{
			UserID:         userInfo.UserID,
			Nickname:       userInfo.Nickname,
			AvatarURL:      userInfo.AvatarURL,
			Coin:           userInfo.Coin,
			InviteCode:     userInfo.InviteCode,
			UsedInviteCode: userInfo.UsedInviteCode,
			LastSignInDate: userInfo.LastSignInDate,
		},
	})
}
