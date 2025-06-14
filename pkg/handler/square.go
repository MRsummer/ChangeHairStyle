package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MRsummer/ChangeHairStyle/pkg/db"
	"github.com/MRsummer/ChangeHairStyle/pkg/model"
	"github.com/gin-gonic/gin"
)

// HandleShareToSquare 处理分享到广场请求
func HandleShareToSquare(c *gin.Context) {
	var req model.ShareToSquareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 分享到广场
	content := &model.SquareContent{
		UserID:   req.UserID,
		RecordID: req.RecordID,
	}

	dbConn := c.MustGet("db").(*sql.DB)
	if err := db.ShareToSquare(dbConn, content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"content_id": content.ID,
		},
	})
}

// HandleGetSquareContents 处理获取广场内容列表请求
func HandleGetSquareContents(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请提供用户ID",
		})
		return
	}

	// 获取分页参数
	cursor := int64(0)
	pageSize := 10
	if cursorStr := c.Query("cursor"); cursorStr != "" {
		if c, err := strconv.ParseInt(cursorStr, 10, 64); err == nil {
			cursor = c
		}
	}
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil {
			pageSize = ps
		}
	}

	// 获取广场内容列表
	dbConn := c.MustGet("db").(*sql.DB)
	response, err := db.GetSquareContents(dbConn, userID, cursor, pageSize)
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
		"data":    response,
	})
}

// HandleLike 处理点赞请求
func HandleLike(c *gin.Context) {
	var req struct {
		UserID    string `json:"user_id" binding:"required"`
		ContentID int64  `json:"content_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	dbConn := c.MustGet("db").(*sql.DB)
	err := db.LikeContent(dbConn, req.UserID, req.ContentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": fmt.Sprintf("点赞操作失败: %v", err),
		})
		return
	}

	// 获取最新的点赞状态
	var isLiked bool
	err = dbConn.QueryRow("SELECT EXISTS(SELECT 1 FROM like_record WHERE user_id = ? AND content_id = ?)",
		req.UserID, req.ContentID).Scan(&isLiked)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": fmt.Sprintf("获取点赞状态失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"is_liked": isLiked,
		},
	})
}
