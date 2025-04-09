package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/logger"
	"github.com/Petr09Mitin/xrust-beze-back/internal/router/middleware"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/credentials/insecure"

	user_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/user"
	grpc_handler "github.com/Petr09Mitin/xrust-beze-back/internal/router/grpc/user"
	http_handler "github.com/Petr09Mitin/xrust-beze-back/internal/router/http/user"
	user_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/user"
	authpb "github.com/Petr09Mitin/xrust-beze-back/proto/auth"
	filepb "github.com/Petr09Mitin/xrust-beze-back/proto/file"
	pb "github.com/Petr09Mitin/xrust-beze-back/proto/user"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

var (
	httpServer *http.Server
	grpcServer *grpc.Server
)

func main() {
	log := logger.NewLogger()
	log.Println("Starting user microservice...")

	cfg, err := config.NewUser()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load user config")
	}

	httpPort := cfg.HTTP.Port
	grpcPort := cfg.GRPC.Port

	db, err := initMongo(log, cfg.Mongo)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init mongo db")
	}

	fileGRPCConn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", cfg.Services.File.Host, cfg.Services.File.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to file_service")
		return
	}
	fileGRPCClient := filepb.NewFileServiceClient(fileGRPCConn)
	defer fileGRPCConn.Close()

	authGRPCConn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", cfg.Services.Auth.Host, cfg.Services.Auth.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to auth_service")
		return
	}
	authClient := authpb.NewAuthServiceClient(authGRPCConn)
	defer authGRPCConn.Close()

	userRepo := user_repo.NewUserRepository(db, 10*time.Second, log)
	userService := user_service.NewUserService(userRepo, fileGRPCClient, authClient, 10*time.Second, log)

	skillRepo := user_repo.NewSkillRepository(db, 10*time.Second, log)
	skillService := user_service.NewSkillService(skillRepo, 10*time.Second, log)

	// Создание каналов для сигналов завершения
	errChan := make(chan error)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запуск HTTP сервера
	go func() {
		router := gin.Default()
		router.Use(middleware.CORSMiddleware())
		http_handler.NewUserHandler(router, userService, authClient)
		http_handler.NewSkillHandler(router, skillService)
		httpServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", httpPort),
			Handler: router,
		}
		log.Printf("HTTP server starting on port %d...", httpPort)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("failed to run HTTP server: %v", err)
		}
	}()

	// Запуск gRPC сервера
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
		if err != nil {
			errChan <- fmt.Errorf("failed to listen on port %d: %v", grpcPort, err)
			return
		}

		grpcServer = grpc.NewServer()
		userGrpcService := grpc_handler.NewUserService(userService, log)
		pb.RegisterUserServiceServer(grpcServer, userGrpcService)

		log.Printf("gRPC server starting on port %d...", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			errChan <- fmt.Errorf("failed to serve gRPC: %v", err)
		}
	}()

	// Ожидание сигнала завершения или ошибки
	select {
	case err := <-errChan:
		log.Printf("Error: %v", err)
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
	}

	log.Println("Shutting down servers...")

	// graceful shutdown для HTTP и gRPC серверов
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to shutdown http server")
	}
	grpcServer.GracefulStop()

	log.Println("User microservice stopped")
}

func initMongo(log zerolog.Logger, cfg *config.Mongo) (*mongo.Database, error) {
	log.Println("Connecting to MongoDB...")

	uri := fmt.Sprintf(
		"mongodb://%s:%s@%s:%d/?authSource=admin",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
	)
	dbName := cfg.Database

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("Connecting to MongoDB with URI: %s", uri)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to MongoDB")
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal().Err(err).Msg("failed to ping MongoDB")
		return nil, err
	}

	log.Println("Connected to MongoDB successfully")
	return client.Database(dbName), nil
}
