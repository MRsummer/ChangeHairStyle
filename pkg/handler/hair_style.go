package handler

import (
	"fmt"
	"net/http"
	"os"

	"github.com/MRsummer/ChangeHairStyle/pkg/cos"
	"github.com/MRsummer/ChangeHairStyle/pkg/volcengine"
	"github.com/gin-gonic/gin"
)

// HairStyleRequest 换发型请求
type HairStyleRequest struct {
	ImageURL    string `json:"image_url"`
	Base64Image string `json:"base64_image"`
	Prompt      string `json:"prompt" binding:"required"`
}

// HairStyleResponse 换发型响应
type HairStyleResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ImageURL string `json:"image_url"`
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

	// 调用火山引擎API
	client := volcengine.NewClient(
		os.Getenv("VOLCENGINE_ACCESS_KEY_ID"),
		os.Getenv("VOLCENGINE_SECRET_ACCESS_KEY"),
	)

	var imageURL string
	var err error

	if req.ImageURL != "" {
		// 直接使用图片URL调用API
		imageURL, err = client.GenerateHairStyle(req.ImageURL, req.Prompt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": fmt.Sprintf("调用火山引擎API失败: %v", err),
			})
			return
		}
	} else if req.Base64Image != "" {
		// 使用base64图片数据调用API
		imageURL, err = client.GenerateHairStyleWithBase64(req.Base64Image, req.Prompt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": fmt.Sprintf("调用火山引擎API失败: %v", err),
			})
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

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"image_url": permanentURL,
		},
	})
} 