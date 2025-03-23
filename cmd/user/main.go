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

	grpc_handler "github.com/Petr09Mitin/xrust-beze-back/internal/router/grpc"
	http_handler "github.com/Petr09Mitin/xrust-beze-back/internal/router/http"
	"github.com/Petr09Mitin/xrust-beze-back/internal/repository/mongodb"
	user_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/user"
	pb "github.com/Petr09Mitin/xrust-beze-back/proto/user"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

func main() {
	// Настройка логирования
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting user microservice...")

	// Получение переменных окружения или использование значений по умолчанию
	mongoURI := getEnv("MONGO_URI", "mongodb://admin:admin@mongo_db:27017/xrust_beze?authSource=admin")
	mongoDBName := getEnv("MONGO_DB_NAME", "xrust_beze")
	httpPort := getEnv("HTTP_PORT", "8080")
	grpcPort := getEnv("GRPC_PORT", "50051")

	// Подключение к MongoDB
	log.Println("Connecting to MongoDB...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err = mongoClient.Disconnect(ctx); err != nil {
			log.Fatalf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	// Проверка соединения с MongoDB
	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB successfully")

	// Инициализация репозитория
	db := mongoClient.Database(mongoDBName)
	userRepo := mongodb.NewUserRepository(db, 10*time.Second)

	// Инициализация сервиса
	userService := user_service.NewUserService(userRepo, 10*time.Second)

	// Создание каналов для сигналов завершения
	errChan := make(chan error)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запуск HTTP сервера
	go func() {
				router := gin.Default()
		http_handler.NewUserHandler(router, userService)

		log.Printf("HTTP server starting on port %s...", httpPort)
		if err := router.Run(":" + httpPort); err != nil && err != http.ErrServerClosed {
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

		s := grpc.NewServer()
		userGrpcService := grpc_handler.NewUserService(userService)
		pb.RegisterUserServiceServer(s, userGrpcService)

		log.Printf("gRPC server starting on port %s...", grpcPort)
		if err := s.Serve(lis); err != nil {
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

	// Здесь можно добавить graceful shutdown для HTTP и gRPC серверов
	// Для HTTP сервера:
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	// if err := httpServer.Shutdown(ctx); err != nil {
	//     log.Printf("HTTP server shutdown error: %v", err)
	// }

	// Для gRPC сервера:
	// grpcServer.GracefulStop()

	log.Println("User microservice stopped")
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}