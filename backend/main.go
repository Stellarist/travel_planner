package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"example.com/travel_planner/backend/handlers"
	"example.com/travel_planner/backend/service"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	api := r.Group("/")
	handlers.RegisterRoutes(api)

	// 加载配置并初始化 Redis
	serverAddr, redisAddr, redisPwd, redisDB := loadConfig()
	if err := service.InitRedis(redisAddr, redisPwd, redisDB); err != nil {
		log.Printf("WARN: Redis init failed: %v (login/register will fail)", err)
	} else {
		log.Printf("Redis connected at %s (db=%d)", redisAddr, redisDB)
	}

	log.Printf("Server starting on %s", serverAddr)
	r.Run(serverAddr)
}

// 读取 config.json，返回服务器地址、redis 配置和全局模型配置
func loadConfig() (serverAddr, redisAddr, redisPwd string, redisDB int) {
	// 默认值
	serverAddr = "127.0.0.1:3000"
	redisAddr = "127.0.0.1:6379"
	redisPwd = ""
	redisDB = 0

	if b, err := os.ReadFile("config.json"); err == nil {
		var cfg service.AppConfig
		if json.Unmarshal(b, &cfg) == nil {
			service.GlobalAppConfig = cfg

			// 读取服务器配置
			if cfg.Server.Host != "" && cfg.Server.Port > 0 {
				serverAddr = fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
			}

			// 读取 Redis 配置
			if cfg.Redis.Addr != "" {
				redisAddr = cfg.Redis.Addr
			}
			if cfg.Redis.Password != "" {
				redisPwd = cfg.Redis.Password
			}
			redisDB = cfg.Redis.DB
		}
	}
	return serverAddr, redisAddr, redisPwd, redisDB
}
