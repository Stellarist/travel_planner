package main

import (
	"log"

	"example.com/travel_planner/backend/api"
	"example.com/travel_planner/backend/config"
	"example.com/travel_planner/backend/handlers"
	"example.com/travel_planner/backend/service"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// 添加CORS中间件
	r.Use(api.CORS())

	apiGroup := r.Group("/")
	handlers.RegisterRoutes(apiGroup)

	// 加载配置并初始化 Redis
	serverAddr, redisAddr, redisPwd, redisDB := config.Load()
	if err := service.InitRedis(redisAddr, redisPwd, redisDB); err != nil {
		log.Printf("WARN: Redis init failed: %v (login/register will fail)", err)
	} else {
		log.Printf("Redis connected at %s (db=%d)", redisAddr, redisDB)
	}

	log.Printf("Server starting on %s", serverAddr)
	r.Run(serverAddr)
}
