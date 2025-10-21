package handlers

import (
	"net/http"

	"example.com/travel_planner/backend/service"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/", RootHandler)
	r.GET("/health", HealthCheckHandler)

	// 认证路由
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/login", LoginHandler)
		authGroup.POST("/register", RegisterHandler)
	}

	// 行程路由（需要认证）
	tripsGroup := r.Group("/api/trips")
	tripsGroup.Use(service.AuthMiddleware())
	{
		tripsGroup.POST("/plan", PlanTripHandler)
		tripsGroup.GET("", GetUserTripsHandler)
		tripsGroup.GET("/:id", GetTripHandler)
		tripsGroup.DELETE("/:id", DeleteTripHandler)
	}
}

func RootHandler(c *gin.Context) {
	c.String(http.StatusOK, "Welcome to Travel Planner Backend API")
}

func HealthCheckHandler(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}
