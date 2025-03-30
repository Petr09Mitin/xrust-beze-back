package main

import (
	"context"
	"fmt"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/logger"
	"github.com/Petr09Mitin/xrust-beze-back/internal/router/middleware"
	"github.com/rs/zerolog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	user_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/user"
	grpc_handler "github.com/Petr09Mitin/xrust-beze-back/internal/router/grpc/user"
	http_handler "github.com/Petr09Mitin/xrust-beze-back/internal/router/http/user"
	user_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/user"
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

	httpPort := getEnv("HTTP_PORT", "8080")
	grpcPort := getEnv("GRPC_PORT", "50051")

	db, err := initMongo(log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init mongo db")
	}

	userRepo := user_repo.NewUserRepository(db, 10*time.Second, log)
	userService := user_service.NewUserService(userRepo, 10*time.Second, log)

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
		http_handler.NewUserHandler(router, userService)
		http_handler.NewSkillHandler(router, skillService)

		httpServer = &http.Server{
			Addr:    ":" + httpPort,
			Handler: router,
		}

		log.Printf("HTTP server starting on port %s...", httpPort)
		// if err := router.Run(":" + httpPort); err != nil && err != httpparser.ErrServerClosed {
		// 	errChan <- fmt.Errorf("failed to run HTTP server: %v", err)
		// }
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("failed to run HTTP server: %v", err)
		}
	}()

	// Запуск gRPC сервера
	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			errChan <- fmt.Errorf("failed to listen on port %s: %v", grpcPort, err)
			return
		}

		grpcServer = grpc.NewServer()
		userGrpcService := grpc_handler.NewUserService(userService, log)
		pb.RegisterUserServiceServer(grpcServer, userGrpcService)

		log.Printf("gRPC server starting on port %s...", grpcPort)
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
		log.Printf("HTTP server shutdown error: %v", err)
	}
	grpcServer.GracefulStop()

	log.Println("User microservice stopped")
}

func initMongo(log zerolog.Logger) (*mongo.Database, error) {
	log.Println("Connecting to MongoDB...")

	uri := getEnv("MONGO_URI", "mongodb://admin:admin@mongo_db:27017/xrust_beze?authSource=admin")
	dbName := getEnv("MONGO_DB_NAME", "xrust_beze")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Err(err)
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Err(err)
		return nil, err
	}

	log.Println("Connected to MongoDB successfully")
	return client.Database(dbName), nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
