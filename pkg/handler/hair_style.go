package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/MRsummer/ChangeHairStyle/pkg/cos"
	"github.com/MRsummer/ChangeHairStyle/pkg/db"
	"github.com/MRsummer/ChangeHairStyle/pkg/logger"
	"github.com/MRsummer/ChangeHairStyle/pkg/middleware"
	"github.com/MRsummer/ChangeHairStyle/pkg/model"
	"github.com/MRsummer/ChangeHairStyle/pkg/volcengine"
	"github.com/gin-gonic/gin"
)

// HairStyleRequest 换发型请求
type HairStyleRequest struct {
	ImageURL    string `json:"image_url"`
	Base64Image string `json:"base64_image"`
	Prompt      string `json:"prompt" binding:"required"`
	UserID      string `json:"user_id" binding:"required"`
}

// HairStyleResponse 换发型响应
type HairStyleResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ImageURL string `json:"image_url"`
		RecordID int64  `json:"record_id"`
	} `json:"data"`
}

// HandleHairStyle 处理换发型请求
func HandleHairStyle(c *gin.Context) {
	var req HairStyleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 检查coin是否足够
	dbConn := c.MustGet("db").(*sql.DB)
	enough, err := db.CheckCoin(dbConn, req.UserID, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}
	if !enough {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    400,
			"message": "造型币不足",
		})
		return
	}

	// 调用火山引擎API
	client := volcengine.NewClient(
		os.Getenv("VOLCENGINE_ACCESS_KEY_ID"),
		os.Getenv("VOLCENGINE_SECRET_ACCESS_KEY"),
	)

	var imageURL string
	if req.ImageURL != "" {
		// 直接使用图片URL调用API
		imageURL, err = client.GenerateHairStyle(req.ImageURL, req.Prompt)
		if err != nil {
			// 检查错误信息是否包含429
			errorMsg := err.Error()
			if strings.Contains(errorMsg, "429") {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "当前使用人数较多、请过5秒后尝试",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": fmt.Sprintf("调用火山引擎API失败: %v", err),
				})
			}
			return
		}
	} else if req.Base64Image != "" {
		// 使用base64图片数据调用API
		imageURL, err = client.GenerateHairStyleWithBase64(req.Base64Image, req.Prompt)
		if err != nil {
			// 检查错误信息是否包含429
			errorMsg := err.Error()
			if strings.Contains(errorMsg, "429") {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "当前使用人数较多、请过5秒后尝试",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": fmt.Sprintf("调用火山引擎API失败: %v", err),
				})
			}
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请提供图片URL或Base64数据",
		})
		return
	}

	// 创建腾讯云 COS 客户端
	cosClient, err := cos.NewClient(
		os.Getenv("COS_SECRET_ID"),
		os.Getenv("COS_SECRET_KEY"),
		os.Getenv("COS_BUCKET"),
		os.Getenv("COS_REGION"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": fmt.Sprintf("创建腾讯云 COS 客户端失败: %v", err),
		})
		return
	}

	// 上传到腾讯云 COS
	permanentURL, err := cosClient.FetchImage(imageURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": fmt.Sprintf("上传到腾讯云 COS 失败: %v", err),
		})
		return
	}

	// 保存生成记录
	record := &model.HairStyleRecord{
		UserID:   req.UserID,
		ImageURL: permanentURL,
		Prompt:   req.Prompt,
	}
	if err := db.SaveHairStyleRecord(dbConn, record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": fmt.Sprintf("保存生成记录失败: %v", err),
		})
		return
	}

	// 扣除coin
	if err := db.DeductCoin(dbConn, req.UserID, 20); err != nil {
		// 记录保存成功但扣除coin失败，记录错误但不影响返回结果
		requestID := middleware.GetRequestID(c)

		//这里需要单独记录，是因为返回了StatusOK，但实际出现了错误
		logger.WithContext(map[string]interface{}{
			"request_id": requestID,
			"user_id":    req.UserID,
		}).WithError(err).Error("扣除coin失败")
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"image_url": permanentURL,
			"record_id": record.ID,
		},
	})
}

// HandleGetRecords 获取用户的生成记录
func HandleGetRecords(c *gin.Context) {
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
		if _, err := fmt.Sscanf(pageStr, "%d", &page); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "页码参数错误",
			})
			return
		}
	}
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if _, err := fmt.Sscanf(pageSizeStr, "%d", &pageSize); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "每页数量参数错误",
			})
			return
		}
	}

	// 获取记录
	dbConn := c.MustGet("db").(*sql.DB)
	response, err := db.GetHairStyleRecords(dbConn, userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": fmt.Sprintf("获取记录失败: %v", err),
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    response,
	})
}
