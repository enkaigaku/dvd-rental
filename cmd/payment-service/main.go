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

	paymentv1 "github.com/tokyoyuan/dvd-rental/gen/proto/payment/v1"
	"github.com/tokyoyuan/dvd-rental/internal/payment/config"
	"github.com/tokyoyuan/dvd-rental/internal/payment/handler"
	"github.com/tokyoyuan/dvd-rental/internal/payment/repository"
	"github.com/tokyoyuan/dvd-rental/internal/payment/service"
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

	// Repository
	paymentRepo := repository.NewPaymentRepository(pool)

	// Service
	paymentSvc := service.NewPaymentService(paymentRepo)

	// Handler
	paymentHandler := handler.NewPaymentHandler(paymentSvc)

	// gRPC server
	grpcServer := grpc.NewServer()
	paymentv1.RegisterPaymentServiceServer(grpcServer, paymentHandler)

	// Health check
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("payment.v1.PaymentService", healthpb.HealthCheckResponse_SERVING)

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
		healthServer.SetServingStatus("payment.v1.PaymentService", healthpb.HealthCheckResponse_NOT_SERVING)
		grpcServer.GracefulStop()
	}()

	log.Printf("payment-service listening on :%s", cfg.GRPCPort)
	return grpcServer.Serve(lis)
}
