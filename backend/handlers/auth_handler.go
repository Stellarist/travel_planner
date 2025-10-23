package handlers

import (
	"context"
	"net/http"
	"time"

	"example.com/travel_planner/backend/api"
	"example.com/travel_planner/backend/service"
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
}

// LoginHandler 处理用户登录
func LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		service.LogWarn("Login request with invalid parameters: %v", err)
		api.RespondError(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	u, err := service.GetUser(ctx, req.Username)
	if err != nil {
		service.LogError("Failed to get user %s: %v", req.Username, err)
		api.RespondError(c, http.StatusInternalServerError, "服务器错误")
		return
	}
	if u == nil || !service.VerifyPassword(req.Password, u.PasswordHash) {
		service.LogWarn("Failed login attempt for user: %s", req.Username)
		api.RespondError(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	token, err := service.GenerateToken(u.ID, u.Username, 24*time.Hour)
	if err != nil {
		service.LogError("Failed to generate token for user %s: %v", req.Username, err)
		api.RespondError(c, http.StatusInternalServerError, "生成 token 失败")
		return
	}

	service.LogInfo("User %s (ID: %d) logged in successfully", req.Username, u.ID)
	c.JSON(http.StatusOK, LoginResponse{
		Success: true,
		Message: "登录成功",
		Token:   token,
		User: &User{
			ID:       u.ID,
			Username: u.Username,
		},
	})
}

// RegisterHandler 处理用户注册
func RegisterHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		service.LogWarn("Register request with invalid parameters: %v", err)
		api.RespondError(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	existing, err := service.GetUser(ctx, req.Username)
	if err != nil {
		service.LogError("Failed to check existing user %s: %v", req.Username, err)
		api.RespondError(c, http.StatusInternalServerError, "服务器错误")
		return
	}
	if existing != nil {
		service.LogWarn("Registration failed - username already exists: %s", req.Username)
		c.JSON(http.StatusConflict, LoginResponse{
			Success: false,
			Message: "用户名已存在",
		})
		return
	}

	userRecord, err := service.CreateUser(ctx, req.Username, req.Password)
	if err != nil {
		service.LogError("Failed to create user %s: %v", req.Username, err)
		api.RespondError(c, http.StatusInternalServerError, "注册失败")
		return
	}

	service.LogInfo("New user registered: %s (ID: %d)", req.Username, userRecord.ID)
	c.JSON(http.StatusCreated, LoginResponse{
		Success: true,
		Message: "注册成功",
	})
}
