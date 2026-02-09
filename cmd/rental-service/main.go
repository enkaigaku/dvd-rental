package main

import (
	"context"
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

	rentalv1 "github.com/tokyoyuan/dvd-rental/gen/proto/rental/v1"
	"github.com/tokyoyuan/dvd-rental/internal/rental/config"
	"github.com/tokyoyuan/dvd-rental/internal/rental/handler"
	"github.com/tokyoyuan/dvd-rental/internal/rental/repository"
	"github.com/tokyoyuan/dvd-rental/internal/rental/service"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return err
	}
	log.Println("connected to database")

	// Repositories
	rentalRepo := repository.NewRentalRepository(pool)
	inventoryRepo := repository.NewInventoryRepository(pool)

	// Services
	rentalSvc := service.NewRentalService(rentalRepo, inventoryRepo)
	inventorySvc := service.NewInventoryService(inventoryRepo, rentalRepo)

	// Handlers
	rentalHandler := handler.NewRentalHandler(rentalSvc)
	inventoryHandler := handler.NewInventoryHandler(inventorySvc)

	// gRPC server
	grpcServer := grpc.NewServer()
	rentalv1.RegisterRentalServiceServer(grpcServer, rentalHandler)
	rentalv1.RegisterInventoryServiceServer(grpcServer, inventoryHandler)

	// Health check
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("rental.v1.RentalService", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("rental.v1.InventoryService", healthpb.HealthCheckResponse_SERVING)

	// Reflection for development tooling
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		return err
	}

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("received signal %v, shutting down gracefully...", sig)
		healthServer.SetServingStatus("rental.v1.RentalService", healthpb.HealthCheckResponse_NOT_SERVING)
		healthServer.SetServingStatus("rental.v1.InventoryService", healthpb.HealthCheckResponse_NOT_SERVING)
		grpcServer.GracefulStop()
	}()

	log.Printf("rental-service listening on :%s", cfg.GRPCPort)
	return grpcServer.Serve(lis)
}
