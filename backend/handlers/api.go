package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/", RootHandler)
	r.GET("/health", HealthCheckHandler)
}

func RootHandler(c *gin.Context) {
	c.String(http.StatusOK, "Welcome to Travel Planner Backend API")
}

func HealthCheckHandler(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}
