package main

import (
	"context"
	"log"

	"backend-go/config"
	"backend-go/controllers"
	"backend-go/database"
	"backend-go/models"
	"backend-go/routes"
	"backend-go/services"

	"github.com/gin-gonic/gin"
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	database.InitDB(cfg)

	// Auto Migrate models
	database.DB.AutoMigrate(
		&models.InterviewSession{},
		&models.InterviewQuestion{},
		&models.InterviewResponse{},
		&models.InterviewFeedback{},
	)

	if cfg.GeminiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY is not set in .env or environment")
	}

	// Initialize services
	aiService, err := services.NewAIService(ctx, cfg.GeminiAPIKey)
	if err != nil {
		log.Fatal("Failed to initialize AI service: ", err)
	}
	cvService := services.NewCVService("./uploads/cv")
	audioService := services.NewAudioService("./uploads/audio", 25*1024*1024)

	// Initialize controllers
	interviewCtrl := controllers.NewInterviewController(aiService, audioService)
	cvCtrl := controllers.NewCVController(cvService)
	audioCtrl := controllers.NewAudioController(aiService, audioService)

	// Initialize Gin router
	r := gin.Default()

	// Register routes
	routes.RegisterRoutes(r, interviewCtrl, cvCtrl, audioCtrl)

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
