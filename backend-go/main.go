package main

import (
	"context"
	"log"
	"time"

	"backend-go/config"
	"backend-go/controllers"
	"backend-go/database"
	"backend-go/models"
	"backend-go/routes"
	"backend-go/services"

	"github.com/gin-contrib/cors"
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
	aiService, err := services.NewAIService(ctx, cfg.GeminiAPIKey, cfg.GeminiModel)
	if err != nil {
		log.Fatal("Failed to initialize AI service: ", err)
	}
	var storage *services.SupabaseStorage
	if cfg.SupabaseURL != "" && cfg.SupabaseKey != "" {
		storage = services.NewSupabaseStorage(cfg.SupabaseURL, cfg.SupabaseKey)
	}

	cvService := services.NewCVService("./uploads/cv", storage, cfg.CVBucket)
	audioService := services.NewAudioService("./uploads/audio", 25*1024*1024, storage, cfg.AudioBucket)

	// Initialize controllers
	interviewCtrl := controllers.NewInterviewController(aiService, audioService)
	cvCtrl := controllers.NewCVController(cvService)
	audioCtrl := controllers.NewAudioController(aiService, audioService)

	// Initialize Gin router
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:60731",
			"http://127.0.0.1:60731",
		},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "apikey", "x-upsert"},
		MaxAge:       12 * time.Hour,
	}))

	// Register routes
	routes.RegisterRoutes(r, interviewCtrl, cvCtrl, audioCtrl)

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
