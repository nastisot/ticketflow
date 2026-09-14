package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	dataURL := os.Getenv("DATABASE_URL")
	if dataURL == "" {
		return nil, errors.New("DATABASE_URL environment variable not set")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	cfg := &Config{
		DatabaseURL: dataURL,
		Port:        port,
	}
	return cfg, nil
}
