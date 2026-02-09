package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	customerv1 "github.com/tokyoyuan/dvd-rental/gen/proto/customer/v1"
	filmv1 "github.com/tokyoyuan/dvd-rental/gen/proto/film/v1"
	paymentv1 "github.com/tokyoyuan/dvd-rental/gen/proto/payment/v1"
	rentalv1 "github.com/tokyoyuan/dvd-rental/gen/proto/rental/v1"
	storev1 "github.com/tokyoyuan/dvd-rental/gen/proto/store/v1"
	"github.com/tokyoyuan/dvd-rental/internal/bff/admin/config"
	"github.com/tokyoyuan/dvd-rental/internal/bff/admin/handler"
	"github.com/tokyoyuan/dvd-rental/internal/bff/admin/router"
	"github.com/tokyoyuan/dvd-rental/pkg/auth"
	"github.com/tokyoyuan/dvd-rental/pkg/grpcutil"
	"github.com/tokyoyuan/dvd-rental/pkg/middleware"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// 1. Load configuration.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// 2. Connect to Redis.
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("parse redis url: %w", err)
	}
	redisClient := redis.NewClient(redisOpts)
	defer redisClient.Close()

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}
	log.Println("connected to Redis")

	// 3. Create auth components.
	jwtManager, err := auth.NewJWTManager(cfg.JWTSecret, cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
	if err != nil {
		return fmt.Errorf("create jwt manager: %w", err)
	}
	refreshStore := auth.NewRefreshTokenStore(redisClient, cfg.RefreshTokenDuration)
	authMw := middleware.NewAuthMiddleware(jwtManager)

	// 4. Create gRPC connections.
	storeConn := grpcutil.MustDial(grpcutil.DefaultClientConfig(cfg.StoreServiceAddr))
	defer storeConn.Close()

	customerConn := grpcutil.MustDial(grpcutil.DefaultClientConfig(cfg.CustomerServiceAddr))
	defer customerConn.Close()

	filmConn := grpcutil.MustDial(grpcutil.DefaultClientConfig(cfg.FilmServiceAddr))
	defer filmConn.Close()

	rentalConn := grpcutil.MustDial(grpcutil.DefaultClientConfig(cfg.RentalServiceAddr))
	defer rentalConn.Close()

	paymentConn := grpcutil.MustDial(grpcutil.DefaultClientConfig(cfg.PaymentServiceAddr))
	defer paymentConn.Close()

	log.Println("gRPC clients initialized")

	// 5. Create gRPC clients.
	storeClient := storev1.NewStoreServiceClient(storeConn)
	staffClient := storev1.NewStaffServiceClient(storeConn)
	customerClient := customerv1.NewCustomerServiceClient(customerConn)
	filmClient := filmv1.NewFilmServiceClient(filmConn)
	actorClient := filmv1.NewActorServiceClient(filmConn)
	categoryClient := filmv1.NewCategoryServiceClient(filmConn)
	languageClient := filmv1.NewLanguageServiceClient(filmConn)
	rentalClient := rentalv1.NewRentalServiceClient(rentalConn)
	inventoryClient := rentalv1.NewInventoryServiceClient(rentalConn)
	paymentClient := paymentv1.NewPaymentServiceClient(paymentConn)

	// 6. Create handlers.
	authHandler := handler.NewAuthHandler(staffClient, jwtManager, refreshStore)
	storeHandler := handler.NewStoreHandler(storeClient)
	staffHandler := handler.NewStaffHandler(staffClient)
	customerHandler := handler.NewCustomerHandler(customerClient)
	filmHandler := handler.NewFilmHandler(filmClient, actorClient, categoryClient, languageClient)
	inventoryHandler := handler.NewInventoryHandler(inventoryClient)
	rentalHandler := handler.NewRentalHandler(rentalClient)
	paymentHandler := handler.NewPaymentHandler(paymentClient)

	// 7. Create router.
	mux := router.NewRouter(
		authHandler,
		storeHandler,
		staffHandler,
		customerHandler,
		filmHandler,
		inventoryHandler,
		rentalHandler,
		paymentHandler,
		authMw,
	)

	// 8. Create HTTP server.
	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 9. Graceful shutdown.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("admin-bff listening on :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	sig := <-sigCh
	log.Printf("received signal %v, shutting down gracefully...", sig)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	log.Println("server stopped gracefully")
	return nil
}
