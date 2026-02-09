// Package config handles configuration loading for the rental service.
package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Config holds the rental service configuration.
type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
	GRPCPort    string `envconfig:"GRPC_PORT" default:"50054"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
