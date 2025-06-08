package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"wx-backend/pkg/volcengine"
)

// HairStyleHandler 发型生成处理器
type HairStyleHandler struct {
	client *volcengine.Client
}

// NewHairStyleHandler 创建发型生成处理器
func NewHairStyleHandler(client *volcengine.Client) *HairStyleHandler {
	return &HairStyleHandler{
		client: client,
	}
}

// Generate 生成发型
func (h *HairStyleHandler) Generate(c *gin.Context) {
	var req struct {
		ImageURL string `json:"image_url" binding:"required"`
		Prompt   string `json:"prompt" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
		})
		return
	}

	// 生成发型图片
	imageURL, err := h.client.GenerateHairStyle(req.ImageURL, req.Prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"image_url": imageURL,
		},
	})
} 