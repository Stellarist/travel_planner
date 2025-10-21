package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
	User    *User  `json:"user,omitempty"`
}

// User 用户信息
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// 模拟用户数据库（实际项目中应该使用真实的数据库）
var users = map[string]string{
	"admin":    "admin123",
	"user":     "password123",
	"testuser": "test123",
}

// LoginHandler 处理用户登录
func LoginHandler(c *gin.Context) {
	var req LoginRequest

	// 绑定并验证请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, LoginResponse{
			Success: false,
			Message: "请求参数错误",
		})
		return
	}

	// 验证用户名和密码
	password, exists := users[req.Username]
	if !exists || password != req.Password {
		c.JSON(http.StatusUnauthorized, LoginResponse{
			Success: false,
			Message: "用户名或密码错误",
		})
		return
	}

	// 登录成功，返回用户信息和token（这里使用简单的token，实际项目应使用JWT）
	token := "token_" + req.Username + "_" + "12345"

	c.JSON(http.StatusOK, LoginResponse{
		Success: true,
		Message: "登录成功",
		Token:   token,
		User: &User{
			ID:       1,
			Username: req.Username,
			Email:    req.Username + "@example.com",
		},
	})
}

// RegisterHandler 处理用户注册
func RegisterHandler(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, LoginResponse{
			Success: false,
			Message: "请求参数错误",
		})
		return
	}

	// 检查用户是否已存在
	if _, exists := users[req.Username]; exists {
		c.JSON(http.StatusConflict, LoginResponse{
			Success: false,
			Message: "用户名已存在",
		})
		return
	}

	// 注册新用户
	users[req.Username] = req.Password

	c.JSON(http.StatusCreated, LoginResponse{
		Success: true,
		Message: "注册成功",
	})
}
