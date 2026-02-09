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

	filmv1 "github.com/enkaigaku/dvd-rental/gen/proto/film/v1"
	"github.com/enkaigaku/dvd-rental/internal/film/config"
	"github.com/enkaigaku/dvd-rental/internal/film/handler"
	"github.com/enkaigaku/dvd-rental/internal/film/repository"
	"github.com/enkaigaku/dvd-rental/internal/film/service"
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
	filmRepo := repository.NewFilmRepository(pool)
	actorRepo := repository.NewActorRepository(pool)
	categoryRepo := repository.NewCategoryRepository(pool)
	languageRepo := repository.NewLanguageRepository(pool)

	// Services
	filmSvc := service.NewFilmService(filmRepo, actorRepo, categoryRepo, languageRepo)
	actorSvc := service.NewActorService(actorRepo)
	categorySvc := service.NewCategoryService(categoryRepo)
	languageSvc := service.NewLanguageService(languageRepo)

	// Handlers
	filmHandler := handler.NewFilmHandler(filmSvc)
	actorHandler := handler.NewActorHandler(actorSvc)
	categoryHandler := handler.NewCategoryHandler(categorySvc)
	languageHandler := handler.NewLanguageHandler(languageSvc)

	// gRPC server
	grpcServer := grpc.NewServer()
	filmv1.RegisterFilmServiceServer(grpcServer, filmHandler)
	filmv1.RegisterActorServiceServer(grpcServer, actorHandler)
	filmv1.RegisterCategoryServiceServer(grpcServer, categoryHandler)
	filmv1.RegisterLanguageServiceServer(grpcServer, languageHandler)

	// Health check
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("film.v1.FilmService", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("film.v1.ActorService", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("film.v1.CategoryService", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("film.v1.LanguageService", healthpb.HealthCheckResponse_SERVING)

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
		healthServer.SetServingStatus("film.v1.FilmService", healthpb.HealthCheckResponse_NOT_SERVING)
		healthServer.SetServingStatus("film.v1.ActorService", healthpb.HealthCheckResponse_NOT_SERVING)
		healthServer.SetServingStatus("film.v1.CategoryService", healthpb.HealthCheckResponse_NOT_SERVING)
		healthServer.SetServingStatus("film.v1.LanguageService", healthpb.HealthCheckResponse_NOT_SERVING)
		grpcServer.GracefulStop()
	}()

	log.Printf("film-service listening on :%s", cfg.GRPCPort)
	return grpcServer.Serve(lis)
}
