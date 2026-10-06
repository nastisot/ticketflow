package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL    string
	Port           string
	BookingTTL     time.Duration
	KafkaBrokers   []string
	KafkaTopic     string
	OutboxInterval time.Duration
	OutboxLimit    int
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

	kafkaBrokersString := strings.TrimSpace(os.Getenv("KAFKA_BROKERS"))
	if kafkaBrokersString == "" {
		kafkaBrokersString = "localhost:9092"
	}
	kafkaBrokersRaw := strings.Split(kafkaBrokersString, ",")

	kafkaBrokers := make([]string, len(kafkaBrokersRaw))
	for _, broker := range kafkaBrokersRaw {
		broker = strings.TrimSpace(broker)
		if broker == "" {
			kafkaBrokers = append(kafkaBrokers, broker)
		}
	}

	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "booking-events"
	}

	outboxIntervalString := os.Getenv("OUTBOX_INTERVAL")
	if outboxIntervalString == "" {
		outboxIntervalString = "5s"
	}
	outboxInterval, err := time.ParseDuration(outboxIntervalString)
	if err != nil {
		return nil, fmt.Errorf("invalid OUTBOX_INTERVAL: %w", err)
	}
	if outboxInterval <= 0 {
		return nil, errors.New("OUTBOX_LIMIT must be greater than 0")
	}

	outboxLimitString := os.Getenv("OUTBOX_LIMIT")
	if outboxLimitString == "" {
		outboxLimitString = "100"
	}
	outboxLimit, err := strconv.Atoi(outboxLimitString)
	if err != nil {
		return nil, fmt.Errorf("invalid OUTBOX_LIMIT: %w", err)
	}
	if outboxLimit <= 0 {
		return nil, errors.New("OUTBOX_LIMIT must be greater than 0")
	}

	cfg := &Config{
		DatabaseURL:    dataURL,
		Port:           port,
		BookingTTL:     bookingTTL,
		KafkaBrokers:   kafkaBrokers,
		KafkaTopic:     kafkaTopic,
		OutboxInterval: outboxInterval,
		OutboxLimit:    outboxLimit,
	}

	return cfg, nil
}
