package main

import (
	"log"

	"example.com/travel_planner/backend/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	api := r.Group("/")
	handlers.RegisterRoutes(api)

	const addr = "127.0.0.1:3000"
	log.Printf("Server starting on %s", addr)
	r.Run(addr)
}
