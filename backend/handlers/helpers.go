package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// getUsernameFromContext 从上下文获取用户名
func getUsernameFromContext(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}
	return username.(string), true
}

// respondUnauthorized 返回未授权响应
func respondUnauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": message})
}

// respondError 返回错误响应
func respondError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{"success": false, "message": message})
}

// respondSuccess 返回成功响应
func respondSuccess(c *gin.Context, data gin.H) {
	response := gin.H{"success": true}
	for k, v := range data {
		response[k] = v
	}
	c.JSON(http.StatusOK, response)
}
