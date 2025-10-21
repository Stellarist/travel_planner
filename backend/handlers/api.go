package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/", RootHandler)
	r.GET("/health", HealthCheckHandler)

	// 认证相关路由
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", LoginHandler)
		auth.POST("/register", RegisterHandler)
	}

	// 行程规划路由
	trips := r.Group("/api/trips")
	{
		trips.POST("/plan", PlanTripHandler)    // 创建行程
		trips.GET("", GetUserTripsHandler)      // 获取用户所有行程
		trips.GET("/:id", GetTripHandler)       // 获取单个行程
		trips.DELETE("/:id", DeleteTripHandler) // 删除行程
	}
}

func RootHandler(c *gin.Context) {
	c.String(http.StatusOK, "Welcome to Travel Planner Backend API")
}

func HealthCheckHandler(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}
