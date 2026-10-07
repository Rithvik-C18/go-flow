package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DB struct {
		ConnString string
	}
	Server struct {
		Port int
	}
	JWT struct {
		Secret          string
		AccessTokenTTL  time.Duration
		RefreshTokenTTL time.Duration
	}
	Integrations struct {
		GeminiAPIKey      string
		GoogleCredentials string
	}
	Execution struct {
		Parallel bool
	}
}

func LoadConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: no .env file found, using environment variables")
	}

	cfg := Config{}

	conn := os.Getenv("DATABASE_URL")
	if conn == "" {
		log.Fatalf("DATABASE_URL env var required")
	}
	cfg.DB.ConnString = conn

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("PORT must be a number, got %q", port)
	}
	cfg.Server.Port = p

	cfg.JWT.Secret = os.Getenv("JWT_SECRET")
	if cfg.JWT.Secret == "" {
		if os.Getenv("GIN_MODE") == "release" {
			log.Fatal("JWT_SECRET is required in production")
		}
		log.Println("Warning: JWT_SECRET not set, using insecure development secret")
		cfg.JWT.Secret = "dev-secret-change-me"
	}

	accessTTL := os.Getenv("JWT_ACCESS_TTL")
	if accessTTL == "" {
		accessTTL = "15m"
	}
	accessDuration, err := time.ParseDuration(accessTTL)
	if err != nil {
		log.Fatalf("JWT_ACCESS_TTL must be a valid duration, got %q", accessTTL)
	}
	cfg.JWT.AccessTokenTTL = accessDuration

	refreshTTL := os.Getenv("JWT_REFRESH_TTL")
	if refreshTTL == "" {
		refreshTTL = "168h"
	}
	refreshDuration, err := time.ParseDuration(refreshTTL)
	if err != nil {
		log.Fatalf("JWT_REFRESH_TTL must be a valid duration, got %q", refreshTTL)
	}
	cfg.JWT.RefreshTokenTTL = refreshDuration

	cfg.Integrations.GeminiAPIKey = os.Getenv("GEMINI_API_KEY")
	if cfg.Integrations.GeminiAPIKey == "" {
		log.Println("Warning: GEMINI_API_KEY not set, Gemini AI node will not work")
	}

	cfg.Integrations.GoogleCredentials = os.Getenv("GOOGLE_CREDENTIALS")
	if cfg.Integrations.GoogleCredentials == "" {
		log.Println("Warning: GOOGLE_CREDENTIALS not set, Google Docs/Excel nodes will not work")
	}

	parallelStr := os.Getenv("EXECUTION_PARALLEL")
	if parallelStr == "" {
		parallelStr = "true"
	}
	cfg.Execution.Parallel = parallelStr == "true" || parallelStr == "1"

	return cfg
}
