package config

import (
	"fmt"
	"os"
)

// Config holds the configuration for the film service.
type Config struct {
	DatabaseURL string
	GRPCPort    string
	LogLevel    string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50052"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	return &Config{
		DatabaseURL: dbURL,
		GRPCPort:    grpcPort,
		LogLevel:    logLevel,
	}, nil
}
