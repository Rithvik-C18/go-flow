package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DB struct {
		ConnString string
	}
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error Loading .env File")
	}

	cfg := Config{}

	conn := os.Getenv("DATABASE_URL")
	if conn == "" {
		log.Fatalf("DATABASE_URL env var required")
	}
	cfg.DB.ConnString = conn
	return cfg
}
