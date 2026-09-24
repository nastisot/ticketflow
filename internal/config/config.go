package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
	BookingTTL  time.Duration
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load .env: %w", err)
	}

	dataURL := os.Getenv("DATABASE_URL")
	if dataURL == "" {
		return nil, errors.New("DATABASE_URL environment variable not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	bookingTTLString := os.Getenv("BOOKING_TTL")
	if bookingTTLString == "" {
		bookingTTLString = "10m"
	}
	bookingTTL, err := time.ParseDuration(bookingTTLString)
	if err != nil {
		return nil, fmt.Errorf("invalid BOOKING_TTL: %w", err)
	}

	cfg := &Config{
		DatabaseURL: dataURL,
		Port:        port,
		BookingTTL:  bookingTTL,
	}
	return cfg, nil
}
