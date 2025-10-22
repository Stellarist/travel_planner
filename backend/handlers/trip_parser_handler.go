package handlers

import (
	"net/http"

	"example.com/travel_planner/backend/api"
	"example.com/travel_planner/backend/service"
	"github.com/gin-gonic/gin"
)

// ParseTextRequest 解析文本请求
type ParseTextRequest struct {
	Text string `json:"text" binding:"required"`
}

// ParseTextResponse 解析文本响应
type ParseTextResponse struct {
	Success bool                    `json:"success"`
	Message string                  `json:"message"`
	Data    *service.ParsedTripInfo `json:"data,omitempty"`
}

// ParseTextHandler 解析语音识别后的文字
func ParseTextHandler(c *gin.Context) {
	var req ParseTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.RespondError(c, http.StatusBadRequest, "请求参数错误: 需要提供text字段")
		return
	}

	if req.Text == "" {
		api.RespondError(c, http.StatusBadRequest, "文本内容不能为空")
		return
	}

	// 解析文本
	parsedInfo := service.ParseTripText(req.Text)

	c.JSON(http.StatusOK, ParseTextResponse{
		Success: true,
		Message: "解析成功",
		Data:    parsedInfo,
	})
}
