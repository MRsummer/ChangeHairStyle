package handler

import (
	"database/sql"
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
	page := 1
	pageSize := 10
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			page = p
		}
	}
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil {
			pageSize = ps
		}
	}

	// 获取广场内容列表
	dbConn := c.MustGet("db").(*sql.DB)
	response, err := db.GetSquareContents(dbConn, userID, page, pageSize)
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

// HandleLikeContent 处理点赞请求
func HandleLikeContent(c *gin.Context) {
	var req model.LikeContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 处理点赞
	dbConn := c.MustGet("db").(*sql.DB)
	if err := db.LikeContent(dbConn, req.UserID, req.ContentID); err != nil {
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
