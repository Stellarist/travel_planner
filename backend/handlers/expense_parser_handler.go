package handlers

import (
	"net/http"

	"example.com/travel_planner/backend/api"
	"example.com/travel_planner/backend/service"
	"github.com/gin-gonic/gin"
)

// ParseExpenseRequest 解析开销查询请求
type ParseExpenseRequest struct {
	Text string `json:"text" binding:"required"`
}

// ParseExpenseResponse 解析开销查询响应
type ParseExpenseResponse struct {
	Success bool                        `json:"success"`
	Message string                      `json:"message"`
	Data    *service.ParsedExpenseQuery `json:"data,omitempty"`
}

// ParseExpenseQueryHandler 解析开销相关的语音文字
func ParseExpenseQueryHandler(c *gin.Context) {
	var req ParseExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		service.LogWarn("Parse expense query failed: invalid request")
		api.RespondError(c, http.StatusBadRequest, "请求参数错误: 需要提供text字段")
		return
	}

	if req.Text == "" {
		service.LogWarn("Parse expense query failed: empty text")
		api.RespondError(c, http.StatusBadRequest, "文本内容不能为空")
		return
	}

	// 解析文本
	parsedInfo := service.ParseExpenseQuery(req.Text)

	service.LogInfo("Parsed expense query: category=%s, dates=%s to %s",
		parsedInfo.Category, parsedInfo.StartDate, parsedInfo.EndDate)
	c.JSON(http.StatusOK, ParseExpenseResponse{
		Success: true,
		Message: "解析成功",
		Data:    parsedInfo,
	})
}
