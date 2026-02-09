// Package grpcutil provides utilities for creating gRPC client connections.
package grpcutil

import (
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ClientConfig holds gRPC client connection configuration.
type ClientConfig struct {
	Address          string
	MaxRecvMsgSize   int
	MaxSendMsgSize   int
	KeepaliveTime    time.Duration
	KeepaliveTimeout time.Duration
}

// DefaultClientConfig returns a client configuration with sensible defaults.
func DefaultClientConfig(address string) ClientConfig {
	return ClientConfig{
		Address:          address,
		MaxRecvMsgSize:   4 * 1024 * 1024, // 4MB
		MaxSendMsgSize:   4 * 1024 * 1024, // 4MB
		KeepaliveTime:    30 * time.Second,
		KeepaliveTimeout: 10 * time.Second,
	}
}

// Dial creates a gRPC client connection with the given configuration.
func Dial(cfg ClientConfig) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(cfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(cfg.MaxRecvMsgSize),
			grpc.MaxCallSendMsgSize(cfg.MaxSendMsgSize),
		),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                cfg.KeepaliveTime,
			Timeout:             cfg.KeepaliveTimeout,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", cfg.Address, err)
	}
	return conn, nil
}

// MustDial creates a gRPC connection or panics. Intended for use in main.go initialization.
func MustDial(cfg ClientConfig) *grpc.ClientConn {
	conn, err := Dial(cfg)
	if err != nil {
		panic(fmt.Sprintf("grpcutil: failed to dial %s: %v", cfg.Address, err))
	}
	return conn
}
