package handlers

import (
	"context"
	"net/http"
	"time"

	"example.com/travel_planner/backend/service"
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
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Trip    *service.TripPlan `json:"trip,omitempty"`
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

	username, ok := getUsernameFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, PlanTripResponse{
			Success: false,
			Message: "未登录",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// 获取用户信息
	user, err := service.GetUser(ctx, username)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, PlanTripResponse{
			Success: false,
			Message: "用户不存在",
		})
		return
	}

	// 构建请求
	tripReq := &service.TripPlanRequest{
		Destination:  req.Destination,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		Budget:       req.Budget,
		Travelers:    req.Travelers,
		Preferences:  req.Preferences,
		SpecialNeeds: req.SpecialNeeds,
	}

	// 生成行程
	plan, err := service.GenerateTripPlan(ctx, tripReq)
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
	if err := service.SaveTripPlan(ctx, plan); err != nil {
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
	username, ok := getUsernameFromContext(c)
	if !ok {
		respondUnauthorized(c, "未登录")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	trips, err := service.GetUserTrips(ctx, username)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "获取行程失败")
		return
	}

	respondSuccess(c, gin.H{"trips": trips})
}

// GetTripHandler 获取单个行程详情
func GetTripHandler(c *gin.Context) {
	tripID := c.Param("id")
	if tripID == "" {
		respondError(c, http.StatusBadRequest, "缺少行程ID")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	trip, err := service.GetTripPlan(ctx, tripID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "获取行程失败")
		return
	}

	if trip == nil {
		respondError(c, http.StatusNotFound, "行程不存在")
		return
	}

	respondSuccess(c, gin.H{"trip": trip})
}

// DeleteTripHandler 删除行程
func DeleteTripHandler(c *gin.Context) {
	tripID := c.Param("id")
	if tripID == "" {
		respondError(c, http.StatusBadRequest, "缺少行程ID")
		return
	}

	username, ok := getUsernameFromContext(c)
	if !ok {
		respondUnauthorized(c, "未登录")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	if err := service.DeleteTripPlan(ctx, tripID, username); err != nil {
		respondError(c, http.StatusInternalServerError, "删除失败")
		return
	}

	respondSuccess(c, gin.H{"message": "删除成功"})
}
