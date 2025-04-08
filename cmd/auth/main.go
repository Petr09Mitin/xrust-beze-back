package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/logger"
	session_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/auth"
	auth_http "github.com/Petr09Mitin/xrust-beze-back/internal/router/http/auth"
	auth_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/auth"
	userpb "github.com/Petr09Mitin/xrust-beze-back/proto/user"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log := logger.NewLogger()
	log.Println("Starting auth microservice...")

	cfg, err := config.NewAuth()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load auth config")
	}

	// Подключение к user service через gRPC
	// userConn, err := grpc.Dial(
	// 	"user_service:50051",
	// 	grpc.WithTransportCredentials(insecure.NewCredentials()),
	// )
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("Failed to connect to user service")
	// }
	// defer userConn.Close()

	// userGRPCConn, err := grpc.NewClient(
	// 	fmt.Sprintf("%s:%d", cfg.Services.UserService.Host, cfg.Services.UserService.Port),
	// 	grpc.WithTransportCredentials(insecure.NewCredentials()),
	// )
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("failed to connect to user_service")
	// 	return
	// }

	userGRPCConn, err := dialWithRetry(fmt.Sprintf("%s:%d", cfg.Services.UserService.Host, cfg.Services.UserService.Port), 5, 2*time.Second)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to user_service after retries")
		return
	}

	defer userGRPCConn.Close()
	userGRPCClient := userpb.NewUserServiceClient(userGRPCConn)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	sessionRepo := session_repo.NewSessionRepository(redisClient, 10*time.Second)
	authService := auth_service.NewAuthService(sessionRepo, userGRPCClient, log, 24*time.Hour)

	// Создание каналов для сигналов завершения
	errChan := make(chan error)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запуск HTTP сервера
	router := gin.Default()
	auth_http.NewAuthHandler(router, authService, cfg)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("failed to run HTTP server: %v", err)
		}
	}()

	log.Printf("HTTP server starting on port %d...", cfg.HTTP.Port)

	// Ожидание сигнала завершения или ошибки
	select {
	case err := <-errChan:
		log.Fatal().Err(err).Msg("Server error")
	case sig := <-sigChan:
		log.Info().Msgf("Received signal %s, shutting down...", sig)
	}
}

func dialWithRetry(address string, maxAttempts int, delay time.Duration) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	var err error
	for i := 0; i < maxAttempts; i++ {
		conn, err = grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			return conn, nil
		}
		time.Sleep(delay)
	}
	return nil, err
}
