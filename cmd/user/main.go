package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	user_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository"
	grpc_handler "github.com/Petr09Mitin/xrust-beze-back/internal/router/grpc"
	http_handler "github.com/Petr09Mitin/xrust-beze-back/internal/router/http"
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
	log.SetFlags(log.LstdFlags | log.Lshortfile) // настройка поведения логгера
	log.Println("Starting user microservice...")

	// mongoURI := getEnv("MONGO_URI", "mongodb://admin:admin@mongo_db:27017/xrust_beze?authSource=admin")
	// mongoDBName := getEnv("MONGO_DB_NAME", "xrust_beze")
	httpPort := getEnv("HTTP_PORT", "8080")
	grpcPort := getEnv("GRPC_PORT", "50051")

	db := initMongo()
	userRepo := user_repo.NewUserRepository(db, 10*time.Second)
	userService := user_service.NewUserService(userRepo, 10*time.Second)

	// Создание каналов для сигналов завершения
	errChan := make(chan error)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запуск HTTP сервера
	go func() {
		router := gin.Default()
		http_handler.NewUserHandler(router, userService)

		httpServer = &http.Server{
			Addr:    ":" + httpPort,
			Handler: router,
		}

		log.Printf("HTTP server starting on port %s...", httpPort)
		// if err := router.Run(":" + httpPort); err != nil && err != http.ErrServerClosed {
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
		userGrpcService := grpc_handler.NewUserService(userService)
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

func initMongo() *mongo.Database {
	log.Println("Connecting to MongoDB...")

	uri := getEnv("MONGO_URI", "mongodb://admin:admin@mongo_db:27017/xrust_beze?authSource=admin")
	dbName := getEnv("MONGO_DB_NAME", "xrust_beze")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	log.Println("Connected to MongoDB successfully")
	return client.Database(dbName)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
