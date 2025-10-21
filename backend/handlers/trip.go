package handlers

import (
	"context"
	"net/http"
	"time"

	"example.com/travel_planner/backend/ai"
	"example.com/travel_planner/backend/store"
	"github.com/gin-gonic/gin"
)

// PlanTripRequest 创建行程请求
type PlanTripRequest struct {
	Destination  string   `json:"destination" binding:"required"`
	StartDate    string   `json:"startDate" binding:"required"`
	EndDate      string   `json:"endDate" binding:"required"`
	Budget       float64  `json:"budget" binding:"required"`
	Travelers    int      `json:"travelers" binding:"required,min=1"`
	Preferences  []string `json:"preferences"`
	SpecialNeeds string   `json:"specialNeeds"`
}

// PlanTripResponse 创建行程响应
type PlanTripResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Trip    *store.TripPlan `json:"trip,omitempty"`
}

// PlanTripHandler 生成行程计划
func PlanTripHandler(c *gin.Context) {
	var req PlanTripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, PlanTripResponse{
			Success: false,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 从 token 获取用户信息
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, PlanTripResponse{
			Success: false,
			Message: "未登录",
		})
		return
	}

	// 去掉 "Bearer " 前缀
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	username, err := store.GetUsernameByToken(ctx, token)
	if err != nil || username == "" {
		c.JSON(http.StatusUnauthorized, PlanTripResponse{
			Success: false,
			Message: "登录已过期",
		})
		return
	}

	// 获取用户信息
	user, err := store.GetUser(ctx, username)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, PlanTripResponse{
			Success: false,
			Message: "用户不存在",
		})
		return
	}

	// 构建请求
	tripReq := &store.TripPlanRequest{
		Destination:  req.Destination,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		Budget:       req.Budget,
		Travelers:    req.Travelers,
		Preferences:  req.Preferences,
		SpecialNeeds: req.SpecialNeeds,
	}

	// 生成行程
	plan, err := ai.GenerateTripPlan(ctx, tripReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, PlanTripResponse{
			Success: false,
			Message: "生成行程失败: " + err.Error(),
		})
		return
	}

	// 关联用户
	plan.UserID = user.ID
	plan.Username = username

	// 保存行程
	if err := store.SaveTripPlan(ctx, plan); err != nil {
		c.JSON(http.StatusInternalServerError, PlanTripResponse{
			Success: false,
			Message: "保存行程失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, PlanTripResponse{
		Success: true,
		Message: "行程规划成功",
		Trip:    plan,
	})
}

// GetUserTripsHandler 获取用户的所有行程
func GetUserTripsHandler(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "未登录"})
		return
	}

	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	username, err := store.GetUsernameByToken(ctx, token)
	if err != nil || username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "登录已过期"})
		return
	}

	trips, err := store.GetUserTrips(ctx, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "获取行程失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"trips":   trips,
	})
}

// GetTripHandler 获取单个行程详情
func GetTripHandler(c *gin.Context) {
	tripID := c.Param("id")
	if tripID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "缺少行程ID"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	trip, err := store.GetTripPlan(ctx, tripID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "获取行程失败"})
		return
	}

	if trip == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "行程不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"trip":    trip,
	})
}

// DeleteTripHandler 删除行程
func DeleteTripHandler(c *gin.Context) {
	tripID := c.Param("id")
	if tripID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "缺少行程ID"})
		return
	}

	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "未登录"})
		return
	}

	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	username, err := store.GetUsernameByToken(ctx, token)
	if err != nil || username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "登录已过期"})
		return
	}

	if err := store.DeleteTripPlan(ctx, tripID, username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "删除成功",
	})
}
