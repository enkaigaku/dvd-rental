// Package main is the entry point for the store service.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	storev1 "github.com/tokyoyuan/dvd-rental/gen/proto/store/v1"
	"github.com/tokyoyuan/dvd-rental/internal/store/config"
	"github.com/tokyoyuan/dvd-rental/internal/store/handler"
	"github.com/tokyoyuan/dvd-rental/internal/store/repository"
	"github.com/tokyoyuan/dvd-rental/internal/store/service"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Load configuration.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Connect to PostgreSQL.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	// Verify database connectivity.
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}
	log.Println("connected to database")

	// Initialize layers: repository → service → handler.
	storeRepo := repository.NewStoreRepository(pool)
	staffRepo := repository.NewStaffRepository(pool)

	storeSvc := service.NewStoreService(storeRepo, staffRepo)
	staffSvc := service.NewStaffService(staffRepo, storeRepo)

	storeHandler := handler.NewStoreHandler(storeSvc)
	staffHandler := handler.NewStaffHandler(staffSvc)

	// Create gRPC server.
	grpcServer := grpc.NewServer()
	storev1.RegisterStoreServiceServer(grpcServer, storeHandler)
	storev1.RegisterStaffServiceServer(grpcServer, staffHandler)

	// Register health check.
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("store.v1.StoreService", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("store.v1.StaffService", healthpb.HealthCheckResponse_SERVING)

	// Register reflection for grpcurl / grpc-client tooling.
	reflection.Register(grpcServer)

	// Start listening.
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		return fmt.Errorf("listen on port %s: %w", cfg.GRPCPort, err)
	}

	// Graceful shutdown on SIGINT/SIGTERM.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("received signal %v, shutting down gracefully...", sig)
		healthServer.SetServingStatus("store.v1.StoreService", healthpb.HealthCheckResponse_NOT_SERVING)
		healthServer.SetServingStatus("store.v1.StaffService", healthpb.HealthCheckResponse_NOT_SERVING)
		grpcServer.GracefulStop()
	}()

	log.Printf("store-service listening on :%s", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("serve grpc: %w", err)
	}

	return nil
}
