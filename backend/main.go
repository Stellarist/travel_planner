package main

import (
	"encoding/json"
	"log"
	"os"

	"example.com/travel_planner/backend/handlers"
	store "example.com/travel_planner/backend/store"
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

	const serverAddr = "127.0.0.1:3000"
	// 初始化 Redis（优先读取 backend/config.json，其次环境变量，最后默认值）
	redisAddr, redisPwd, redisDB := loadRedisConfig()
	if err := store.Init(redisAddr, redisPwd, redisDB); err != nil {
		log.Printf("WARN: Redis init failed: %v (login/register will fail)", err)
	} else {
		log.Printf("Redis connected at %s (db=%d)", redisAddr, redisDB)
	}

	log.Printf("Server starting on %s", serverAddr)
	r.Run(serverAddr)
}

// demo 用户已移除，如需种子数据可自行调用 store.CreateUser 在初始化时创建。

// 配置结构
type redisConfig struct {
	Redis struct {
		Addr     string `json:"addr"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	} `json:"redis"`
}

// 从 backend/config.json 或环境变量读取 Redis 配置
func loadRedisConfig() (addr, pwd string, db int) {
	// 默认值
	addr = "127.0.0.1:6379"
	pwd = ""
	db = 0

	// 只读取本地文件 config.json（运行目录应为 backend）
	if b, err := os.ReadFile("config.json"); err == nil {
		var cfg redisConfig
		if json.Unmarshal(b, &cfg) == nil {
			if cfg.Redis.Addr != "" {
				addr = cfg.Redis.Addr
			}
			if cfg.Redis.Password != "" {
				pwd = cfg.Redis.Password
			}
			// db 即便为 0 也是有效值
			db = cfg.Redis.DB
		}
	}
	return addr, pwd, db
}
