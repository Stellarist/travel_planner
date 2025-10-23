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

	authGroup := r.Group("/api/auth")
	authGroup.POST("/login", LoginHandler)
	authGroup.POST("/register", RegisterHandler)

	tripsGroup := r.Group("/api/trips")
	tripsGroup.Use(service.AuthMiddleware())
	tripsGroup.POST("/plan", PlanTripHandler)
	tripsGroup.GET("", GetUserTripsHandler)
	tripsGroup.GET("/:id", GetTripHandler)
	tripsGroup.DELETE("/:id", DeleteTripHandler)
	tripsGroup.GET("/favorites/list", GetFavoriteTripHandler)      // 获取收藏列表
	tripsGroup.POST("/favorites/:id", AddFavoriteTripHandler)      // 添加收藏
	tripsGroup.DELETE("/favorites/:id", RemoveFavoriteTripHandler) // 取消收藏

	expenseGroup := r.Group("/api/expenses")
	expenseGroup.Use(service.AuthMiddleware())
	expenseGroup.POST("", CreateExpenseHandler)
	expenseGroup.GET("", ListExpensesHandler)
	expenseGroup.GET("/analyze", AnalyzeExpensesHandler)

	exploreGroup := r.Group("/api/favorites")
	exploreGroup.Use(service.AuthMiddleware())
	exploreGroup.GET("", GetFavorites)
	exploreGroup.POST("", AddFavorite)
	exploreGroup.DELETE("/:id", RemoveFavorite)

	parserGroup := r.Group("/api/parser")
	parserGroup.Use(service.AuthMiddleware())
	parserGroup.POST("/parse", ParseTextHandler)
	parserGroup.POST("/parse-expense", ParseExpenseQueryHandler)

}

func RootHandler(c *gin.Context) {
	c.String(http.StatusOK, "Welcome to Travel Planner Backend API")
}

func HealthCheckHandler(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}
