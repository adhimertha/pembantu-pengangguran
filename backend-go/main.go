package main

import (
	"log"

	"backend-go/config"
	"backend-go/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize Gin router
	r := gin.Default()

	// Register routes
	routes.RegisterRoutes(r)

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
