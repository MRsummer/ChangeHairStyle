package handler

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/MRsummer/ChangeHairStyle/pkg/volcengine"
	"github.com/MRsummer/ChangeHairStyle/pkg/qiniu"
)

// HairStyleHandler 发型生成处理器
type HairStyleHandler struct {
	client *volcengine.Client
	qiniu  *qiniu.Client
}

// NewHairStyleHandler 创建发型生成处理器
func NewHairStyleHandler(client *volcengine.Client) *HairStyleHandler {
	// 创建七牛云客户端
	qiniuClient := qiniu.NewClient(
		os.Getenv("QINIU_ACCESS_KEY"),
		os.Getenv("QINIU_SECRET_KEY"),
		os.Getenv("QINIU_BUCKET"),
		os.Getenv("QINIU_DOMAIN"),
	)

	return &HairStyleHandler{
		client: client,
		qiniu:  qiniuClient,
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

	// 转存到七牛云
	permanentURL, err := h.qiniu.FetchImage(imageURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "转存图片失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"image_url": permanentURL,
		},
	})
}

// GenerateWithBase64 使用base64编码的图片数据生成发型
func (h *HairStyleHandler) GenerateWithBase64(c *gin.Context) {
	var req struct {
		Base64Image string `json:"base64_image" binding:"required"`
		Prompt      string `json:"prompt" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
		})
		return
	}

	// 生成发型图片
	imageURL, err := h.client.GenerateHairStyleWithBase64(req.Base64Image, req.Prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	// 转存到七牛云
	permanentURL, err := h.qiniu.FetchImage(imageURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "转存图片失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"image_url": permanentURL,
		},
	})
} 