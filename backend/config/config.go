package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type ModelConfig struct {
	ApiKey  string `json:"apikey"`
	BaseURL string `json:"baseurl"`
	Model   string `json:"model"`
}

type AppConfig struct {
	Server struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"server"`
	Redis struct {
		Addr     string `json:"addr"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	} `json:"redis"`
	Model ModelConfig `json:"model"`
}

var Global AppConfig

// Load 读取 config.json，返回服务器地址、redis 配置
func Load() (serverAddr, redisAddr, redisPwd string, redisDB int) {
	// 默认值
	serverAddr = "127.0.0.1:3000"
	redisAddr = "127.0.0.1:6379"
	redisPwd = ""
	redisDB = 0

	if b, err := os.ReadFile("config.json"); err == nil {
		var cfg AppConfig
		if json.Unmarshal(b, &cfg) == nil {
			Global = cfg

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
