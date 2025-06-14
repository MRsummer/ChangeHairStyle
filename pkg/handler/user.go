package handler

import (
	"database/sql"
	"net/http"

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
