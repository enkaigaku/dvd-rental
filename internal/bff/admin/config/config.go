// Package config handles configuration loading for the admin-bff service.
package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds configuration for the admin-bff service.
type Config struct {
	HTTPPort string `envconfig:"HTTP_PORT" default:"8081"`

	// gRPC backend addresses.
	StoreServiceAddr    string `envconfig:"GRPC_STORE_ADDR" default:"localhost:50051"`
	CustomerServiceAddr string `envconfig:"GRPC_CUSTOMER_ADDR" default:"localhost:50053"`
	FilmServiceAddr     string `envconfig:"GRPC_FILM_ADDR" default:"localhost:50052"`
	RentalServiceAddr   string `envconfig:"GRPC_RENTAL_ADDR" default:"localhost:50054"`
	PaymentServiceAddr  string `envconfig:"GRPC_PAYMENT_ADDR" default:"localhost:50055"`

	// JWT settings.
	JWTSecret            string        `envconfig:"JWT_SECRET" required:"true"`
	AccessTokenDuration  time.Duration `envconfig:"JWT_ACCESS_DURATION" default:"15m"`
	RefreshTokenDuration time.Duration `envconfig:"JWT_REFRESH_DURATION" default:"168h"`

	// Redis URL for refresh token storage.
	RedisURL string `envconfig:"REDIS_URL" required:"true"`

	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
