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

	customerv1 "github.com/enkaigaku/dvd-rental/gen/proto/customer/v1"
	"github.com/enkaigaku/dvd-rental/internal/customer/config"
	"github.com/enkaigaku/dvd-rental/internal/customer/handler"
	"github.com/enkaigaku/dvd-rental/internal/customer/repository"
	"github.com/enkaigaku/dvd-rental/internal/customer/service"
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
	customerRepo := repository.NewCustomerRepository(pool)
	addressRepo := repository.NewAddressRepository(pool)
	cityRepo := repository.NewCityRepository(pool)
	countryRepo := repository.NewCountryRepository(pool)

	// Services
	customerSvc := service.NewCustomerService(customerRepo, addressRepo, cityRepo, countryRepo)
	addressSvc := service.NewAddressService(addressRepo, cityRepo)
	citySvc := service.NewCityService(cityRepo)
	countrySvc := service.NewCountryService(countryRepo)

	// Handlers
	customerHandler := handler.NewCustomerHandler(customerSvc)
	addressHandler := handler.NewAddressHandler(addressSvc)
	cityHandler := handler.NewCityHandler(citySvc)
	countryHandler := handler.NewCountryHandler(countrySvc)

	// gRPC server
	grpcServer := grpc.NewServer()
	customerv1.RegisterCustomerServiceServer(grpcServer, customerHandler)
	customerv1.RegisterAddressServiceServer(grpcServer, addressHandler)
	customerv1.RegisterCityServiceServer(grpcServer, cityHandler)
	customerv1.RegisterCountryServiceServer(grpcServer, countryHandler)

	// Health check
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("customer.v1.CustomerService", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("customer.v1.AddressService", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("customer.v1.CityService", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("customer.v1.CountryService", healthpb.HealthCheckResponse_SERVING)

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
		healthServer.SetServingStatus("customer.v1.CustomerService", healthpb.HealthCheckResponse_NOT_SERVING)
		healthServer.SetServingStatus("customer.v1.AddressService", healthpb.HealthCheckResponse_NOT_SERVING)
		healthServer.SetServingStatus("customer.v1.CityService", healthpb.HealthCheckResponse_NOT_SERVING)
		healthServer.SetServingStatus("customer.v1.CountryService", healthpb.HealthCheckResponse_NOT_SERVING)
		grpcServer.GracefulStop()
	}()

	log.Printf("customer-service listening on :%s", cfg.GRPCPort)
	return grpcServer.Serve(lis)
}
