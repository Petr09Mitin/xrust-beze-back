package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/logger"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/validation"

	session_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/auth"
	grpc_handler "github.com/Petr09Mitin/xrust-beze-back/internal/router/grpc/auth"
	auth_http "github.com/Petr09Mitin/xrust-beze-back/internal/router/http/auth"
	auth_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/auth"

	authpb "github.com/Petr09Mitin/xrust-beze-back/proto/auth"
	userpb "github.com/Petr09Mitin/xrust-beze-back/proto/user"
)

var (
	httpServer *http.Server
	grpcServer *grpc.Server
)

func main() {
	log := logger.NewLogger()
	log.Println("Starting auth microservice...")

	err := validation.Init()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize validators")
	}

	cfg, err := config.NewAuth()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load auth config")
	}

	userGRPCConn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", cfg.Services.UserService.Host, cfg.Services.UserService.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to user_service")
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

	errChan := make(chan error)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		router := gin.Default()
		auth_http.NewAuthHandler(router, authService, cfg, log)

		httpServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.HTTP.Port),
			Handler: router,
		}

		log.Printf("HTTP server starting on port %d...", cfg.HTTP.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("failed to run HTTP server: %v", err)
		}
	}()

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
		if err != nil {
			errChan <- fmt.Errorf("failed to listen on port %d: %v", cfg.GRPC.Port, err)
			return
		}

		grpcServer = grpc.NewServer()
		authGrpcService := grpc_handler.NewAuthService(authService, log)
		authpb.RegisterAuthServiceServer(grpcServer, authGrpcService)

		log.Printf("gRPC server starting on port %d...", cfg.GRPC.Port)
		if err := grpcServer.Serve(lis); err != nil {
			errChan <- fmt.Errorf("failed to serve gRPC: %v", err)
		}
	}()

	select {
	case err := <-errChan:
		log.Fatal().Err(err).Msg("Server error")
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
	}

	log.Println("Shutting down auth microservice...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to shutdown http server")
	}
	defer cancel()
	grpcServer.GracefulStop()

	log.Println("Auth microservice stopped")
}

// package main

// import (
// 	"fmt"
// 	"net"
// 	"net/http"
// 	"os"
// 	"os/signal"
// 	"syscall"
// 	"time"

// 	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
// 	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/logger"
// 	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/validation"
// 	session_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/auth"
// 	grpc_handler "github.com/Petr09Mitin/xrust-beze-back/internal/router/grpc/auth"
// 	auth_http "github.com/Petr09Mitin/xrust-beze-back/internal/router/http/auth"
// 	auth_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/auth"
// 	authpb "github.com/Petr09Mitin/xrust-beze-back/proto/auth"
// 	userpb "github.com/Petr09Mitin/xrust-beze-back/proto/user"
// 	"github.com/gin-gonic/gin"
// 	"github.com/redis/go-redis/v9"
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/credentials/insecure"
// )

// var grpcServer *grpc.Server

// func main() {
// 	log := logger.NewLogger()
// 	log.Println("Starting auth microservice...")

// 	validation.Init()

// 	cfg, err := config.NewAuth()
// 	if err != nil {
// 		log.Fatal().Err(err).Msg("Failed to load auth config")
// 	}

// 	userGRPCConn, err := dialWithRetry(fmt.Sprintf("%s:%d", cfg.Services.UserService.Host, cfg.Services.UserService.Port), 5, 2*time.Second)
// 	if err != nil {
// 		log.Fatal().Err(err).Msg("Failed to connect to user_service after retries")
// 		return
// 	}
// 	userGRPCClient := userpb.NewUserServiceClient(userGRPCConn)
// 	defer userGRPCConn.Close()

// 	redisClient := redis.NewClient(&redis.Options{
// 		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
// 		Password: cfg.Redis.Password,
// 		DB:       cfg.Redis.DB,
// 	})

// 	sessionRepo := session_repo.NewSessionRepository(redisClient, 10*time.Second)
// 	authService := auth_service.NewAuthService(sessionRepo, userGRPCClient, log, 24*time.Hour)

// 	errChan := make(chan error)
// 	sigChan := make(chan os.Signal, 1)
// 	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

// 	// Запуск HTTP сервера
// 	router := gin.Default()
// 	auth_http.NewAuthHandler(router, authService, cfg)

// 	server := &http.Server{
// 		Addr:    fmt.Sprintf(":%d", cfg.HTTP.Port),
// 		Handler: router,
// 	}

// 	// Запуск http
// 	go func() {
// 		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
// 			errChan <- fmt.Errorf("failed to run HTTP server: %v", err)
// 		}
// 	}()

// 	// Запуск пкзс
// 	go func() {
// 		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
// 		if err != nil {
// 			errChan <- fmt.Errorf("failed to listen on port %d: %v", cfg.GRPC.Port, err)
// 			return
// 		}

// 		grpcServer = grpc.NewServer()
// 		authGrpcService := grpc_handler.NewAuthService(authService, log)
// 		authpb.RegisterAuthServiceServer(grpcServer, authGrpcService)

// 		log.Printf("gRPC server starting on port %d...", cfg.GRPC.Port)
// 		if err := grpcServer.Serve(lis); err != nil {
// 			errChan <- fmt.Errorf("failed to serve gRPC: %v", err)
// 		}
// 	}()

// 	log.Printf("HTTP server starting on port %d...", cfg.HTTP.Port)

// 	// Ожидание сигнала завершения или ошибки
// 	select {
// 	case err := <-errChan:
// 		log.Fatal().Err(err).Msg("Server error")
// 	case sig := <-sigChan:
// 		log.Info().Msgf("Received signal %s, shutting down...", sig)
// 	}
// }

// func dialWithRetry(address string, maxAttempts int, delay time.Duration) (*grpc.ClientConn, error) {
// 	var conn *grpc.ClientConn
// 	var err error
// 	for i := 0; i < maxAttempts; i++ {
// 		// conn, err = grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
// 		conn, err = grpc.Dial(
// 			address,
// 			grpc.WithTransportCredentials(insecure.NewCredentials()),
// 			grpc.WithTimeout(5*time.Second),
// 		)
// 		if err == nil {
// 			return conn, nil
// 		}
// 		time.Sleep(delay)
// 	}
// 	return nil, err
// }

// // Подключение к user service через gRPC
// // userConn, err := grpc.Dial(
// // 	"user_service:50051",
// // 	grpc.WithTransportCredentials(insecure.NewCredentials()),
// // )
// // if err != nil {
// // 	log.Fatal().Err(err).Msg("Failed to connect to user service")
// // }
// // defer userConn.Close()

// // userGRPCConn, err := grpc.NewClient(
// // 	fmt.Sprintf("%s:%d", cfg.Services.UserService.Host, cfg.Services.UserService.Port),
// // 	grpc.WithTransportCredentials(insecure.NewCredentials()),
// // )
// // if err != nil {
// // 	log.Fatal().Err(err).Msg("failed to connect to user_service")
// // 	return
// // }
