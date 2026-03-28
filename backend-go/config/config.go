package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	GeminiAPIKey string
	GeminiModel  string
	DatabaseURL  string
	SupabaseURL  string
	SupabaseKey  string
	CVBucket     string
	AudioBucket  string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		Port:         getEnv("PORT", "8080"),
		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),
		GeminiModel:  getEnv("GEMINI_MODEL", "gemini-1.5-flash-001"),
		DatabaseURL:  getEnv("DATABASE_URL", ""),
		SupabaseURL:  getEnv("SUPABASE_URL", ""),
		SupabaseKey:  getEnv("SUPABASE_SERVICE_ROLE_KEY", ""),
		CVBucket:     getEnv("SUPABASE_BUCKET_CV", "cv"),
		AudioBucket:  getEnv("SUPABASE_BUCKET_AUDIO", "audio"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
