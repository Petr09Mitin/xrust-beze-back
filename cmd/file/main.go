package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/logger"
	filerepo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/file"
	filegrpc "github.com/Petr09Mitin/xrust-beze-back/internal/router/grpc/file"
	filehandler "github.com/Petr09Mitin/xrust-beze-back/internal/router/http/file"
	"github.com/Petr09Mitin/xrust-beze-back/internal/router/middleware"
	"github.com/Petr09Mitin/xrust-beze-back/internal/services/file"
	pb "github.com/Petr09Mitin/xrust-beze-back/proto/file"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var httpServer *http.Server
	var grpcServer *grpc.Server

	log := logger.NewLogger()
	log.Println("Starting user microservice...")

	cfg, err := config.NewFile()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load user config")
	}

	minioClient, err := minio.New(cfg.Minio.Endpoint, &minio.Options{
		Secure: cfg.Minio.UseSSL,
		Creds:  credentials.NewStaticV4(cfg.Minio.Username, cfg.Minio.Password, ""),
	},
	)

	httpPort := cfg.HTTP.Port
	grpcPort := cfg.GRPC.Port

	fileRepo, err := filerepo.NewFileRepo(minioClient, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize file repo")
	}
	fileService := file.NewFileService(fileRepo, log)

	// Создание каналов для сигналов завершения
	errChan := make(chan error)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запуск HTTP сервера
	go func() {
		router := gin.Default()
		router.Use(middleware.CORSMiddleware())
		filehandler.NewFileHandler(router, fileService, log)
		log.Printf("HTTP server starting on port %s...", httpPort)
		httpServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", httpPort),
			Handler: router.Handler(),
		}

		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("failed to run HTTP server: %v", err)
		}
	}()

	// Запуск gRPC сервера
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
		if err != nil {
			errChan <- fmt.Errorf("failed to listen on port %s: %v", grpcPort, err)
			return
		}

		grpcServer = grpc.NewServer()
		fileGrpcService := filegrpc.NewFileGRPCService(fileService, log)
		pb.RegisterFileServiceServer(grpcServer, fileGrpcService)

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
		log.Fatal().Err(err).Msg("failed to shutdown http server")
	}
	grpcServer.GracefulStop()

	log.Println("File microservice stopped")
}
