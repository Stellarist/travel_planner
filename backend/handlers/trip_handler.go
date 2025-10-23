package handlers

import (
	"context"
	"net/http"
	"time"

	"example.com/travel_planner/backend/api"
	"example.com/travel_planner/backend/service"
	"github.com/gin-gonic/gin"
)

// TripRequest 创建行程请求
type TripRequest struct {
	Destination  string   `json:"destination" binding:"required"`
	StartDate    string   `json:"startDate" binding:"required"`
	EndDate      string   `json:"endDate" binding:"required"`
	Budget       float64  `json:"budget"`
	Travelers    int      `json:"travelers" binding:"required,min=1"`
	Preferences  []string `json:"preferences"`
	SpecialNeeds string   `json:"specialNeeds"`
}

// TripResponse 创建行程响应
type TripResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Trip    *service.TripPlan `json:"trip,omitempty"`
}

// PlanTripHandler 生成行程计划
func PlanTripHandler(c *gin.Context) {
	var req TripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.RespondError(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	if req.Budget <= 0 {
		req.Budget = 2000
	}

	username, ok := api.GetUsername(c)
	if !ok {
		api.RespondError(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 增加超时时间到 2 分钟，因为 AI 生成行程需要较长时间
	ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
	defer cancel()

	user, err := service.GetUser(ctx, username)
	if err != nil || user == nil {
		api.RespondError(c, http.StatusUnauthorized, "用户不存在")
		return
	}

	tripReq := &service.TripPlanRequest{
		Destination:  req.Destination,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		Budget:       req.Budget,
		Travelers:    req.Travelers,
		Preferences:  req.Preferences,
		SpecialNeeds: req.SpecialNeeds,
	}

	plan, err := service.GenerateTripPlan(ctx, tripReq)
	if err != nil {
		api.RespondError(c, http.StatusInternalServerError, "生成行程失败: "+err.Error())
		return
	}

	plan.UserID = user.ID
	plan.Username = username

	if err := service.SaveTripPlan(ctx, plan); err != nil {
		api.RespondError(c, http.StatusInternalServerError, "保存行程失败")
		return
	}

	c.JSON(http.StatusOK, TripResponse{
		Success: true,
		Message: "行程规划成功",
		Trip:    plan,
	})
}

// GetUserTripsHandler 获取用户的所有行程
func GetUserTripsHandler(c *gin.Context) {
	username, ok := api.GetUsername(c)
	if !ok {
		api.RespondError(c, http.StatusUnauthorized, "未登录")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	trips, err := service.GetUserTrips(ctx, username)
	if err != nil {
		api.RespondError(c, http.StatusInternalServerError, "获取行程失败")
		return
	}

	api.RespondSuccess(c, trips)
}

// GetTripHandler 获取单个行程详情
func GetTripHandler(c *gin.Context) {
	tripID := c.Param("id")
	if tripID == "" {
		api.RespondError(c, http.StatusBadRequest, "缺少行程ID")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	trip, err := service.GetTripPlan(ctx, tripID)
	if err != nil {
		api.RespondError(c, http.StatusInternalServerError, "获取行程失败")
		return
	}
	if trip == nil {
		api.RespondError(c, http.StatusNotFound, "行程不存在")
		return
	}

	api.RespondSuccess(c, trip)
}

// DeleteTripHandler 删除行程
func DeleteTripHandler(c *gin.Context) {
	tripID := c.Param("id")
	username, ok := api.GetUsername(c)
	if !ok {
		api.RespondError(c, http.StatusUnauthorized, "未登录")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	if err := service.DeleteTripPlan(ctx, tripID, username); err != nil {
		api.RespondError(c, http.StatusInternalServerError, "删除失败")
		return
	}

	api.RespondSuccess(c, gin.H{"message": "删除成功"})
}

// GetFavoriteTripHandler 获取收藏的行程列表
func GetFavoriteTripHandler(c *gin.Context) {
	username, ok := api.GetUsername(c)
	if !ok {
		api.RespondError(c, http.StatusUnauthorized, "未登录")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	trips, err := service.GetUserFavoriteTrips(ctx, username)
	if err != nil {
		api.RespondError(c, http.StatusInternalServerError, "获取收藏失败")
		return
	}

	api.RespondSuccess(c, trips)
}

// AddFavoriteTripHandler 添加行程到收藏
func AddFavoriteTripHandler(c *gin.Context) {
	username, ok := api.GetUsername(c)
	if !ok {
		api.RespondError(c, http.StatusUnauthorized, "未登录")
		return
	}

	tripID := c.Param("id")
	if tripID == "" {
		api.RespondError(c, http.StatusBadRequest, "缺少行程ID")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	if err := service.AddFavoriteTrip(ctx, username, tripID); err != nil {
		api.RespondError(c, http.StatusInternalServerError, "收藏失败: "+err.Error())
		return
	}

	api.RespondSuccess(c, gin.H{"message": "收藏成功"})
}

// RemoveFavoriteTripHandler 取消收藏行程
func RemoveFavoriteTripHandler(c *gin.Context) {
	username, ok := api.GetUsername(c)
	if !ok {
		api.RespondError(c, http.StatusUnauthorized, "未登录")
		return
	}

	tripID := c.Param("id")
	if tripID == "" {
		api.RespondError(c, http.StatusBadRequest, "缺少行程ID")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	if err := service.RemoveFavoriteTrip(ctx, username, tripID); err != nil {
		api.RespondError(c, http.StatusInternalServerError, "取消收藏失败")
		return
	}

	api.RespondSuccess(c, gin.H{"message": "已取消收藏"})
}
