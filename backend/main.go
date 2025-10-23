package main

import (
	"example.com/travel_planner/backend/api"
	"example.com/travel_planner/backend/config"
	"example.com/travel_planner/backend/handlers"
	"example.com/travel_planner/backend/service"
	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化日志系统
	if err := service.InitLogger(); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer service.CloseLogger()

	r := gin.Default()

	// 添加CORS中间件
	r.Use(api.CORS())

	apiGroup := r.Group("/")
	handlers.RegisterRoutes(apiGroup)

	// 加载配置并初始化 Redis
	serverAddr, redisAddr, redisPwd, redisDB := config.Load()
	if err := service.InitRedis(redisAddr, redisPwd, redisDB); err != nil {
		service.LogError("Redis init failed: %v (login/register will fail)", err)
	} else {
		service.LogInfo("Redis connected at %s (db=%d)", redisAddr, redisDB)
	}

	service.LogInfo("Server starting on %s", serverAddr)
	r.Run(serverAddr)
}
