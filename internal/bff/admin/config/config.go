package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds configuration for the admin-bff service.
type Config struct {
	HTTPPort string

	// gRPC backend addresses.
	StoreServiceAddr    string
	CustomerServiceAddr string
	FilmServiceAddr     string
	RentalServiceAddr   string
	PaymentServiceAddr  string

	// JWT settings.
	JWTSecret            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration

	// Redis URL for refresh token storage.
	RedisURL string

	LogLevel string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL is required")
	}

	accessDur := parseDuration(os.Getenv("JWT_ACCESS_DURATION"), 15*time.Minute)
	refreshDur := parseDuration(os.Getenv("JWT_REFRESH_DURATION"), 168*time.Hour)

	return &Config{
		HTTPPort:             envOrDefault("HTTP_PORT", "8081"),
		StoreServiceAddr:     envOrDefault("GRPC_STORE_ADDR", "localhost:50051"),
		CustomerServiceAddr:  envOrDefault("GRPC_CUSTOMER_ADDR", "localhost:50053"),
		FilmServiceAddr:      envOrDefault("GRPC_FILM_ADDR", "localhost:50052"),
		RentalServiceAddr:    envOrDefault("GRPC_RENTAL_ADDR", "localhost:50054"),
		PaymentServiceAddr:   envOrDefault("GRPC_PAYMENT_ADDR", "localhost:50055"),
		JWTSecret:            jwtSecret,
		AccessTokenDuration:  accessDur,
		RefreshTokenDuration: refreshDur,
		RedisURL:             redisURL,
		LogLevel:             envOrDefault("LOG_LEVEL", "info"),
	}, nil
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func parseDuration(s string, fallback time.Duration) time.Duration {
	if s == "" {
		return fallback
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return fallback
	}
	return d
}
