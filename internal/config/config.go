package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds runtime settings loaded from the environment.
type Config struct {
	MaxRequests int
	Window      time.Duration
	Addr        string
}

// LoadConfig reads settings from the environment (defaults match the assignment: 5 / minute).
// Loads `.env` from the process working directory via godotenv.
func LoadConfig() Config {
	_ = godotenv.Load()

	cfg := Config{
		Addr: getEnv("SERVER_ADDR", ":8080"),
	}

	if n, err := strconv.Atoi(getEnv("RATE_LIMIT_MAX_REQUESTS", "5")); err == nil && n > 0 {
		cfg.MaxRequests = n
	} else {
		log.Printf("invalid RATE_LIMIT_MAX_REQUESTS, using default 5")
		cfg.MaxRequests = 5
	}

	if d, err := time.ParseDuration(getEnv("RATE_LIMIT_WINDOW", "1m")); err == nil && d > 0 {
		cfg.Window = d
	} else {
		log.Printf("invalid RATE_LIMIT_WINDOW (use e.g. 1m, 60s), using default 1m")
		cfg.Window = time.Minute
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
