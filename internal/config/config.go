package config

import (
	"log"
	"os"
	"strconv"

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
		Secret string
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
		log.Println("Warning: JWT_SECRET not set, using insecure development secret")
		cfg.JWT.Secret = "dev-secret-change-me"
	}

	return cfg
}
