package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/Petr09Mitin/xrust-beze-back/internal/repository/rag_client"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/logger"
	study_material_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/study_material"
	study_material_http "github.com/Petr09Mitin/xrust-beze-back/internal/router/http/study_material"
	"github.com/Petr09Mitin/xrust-beze-back/internal/router/middleware"
	study_material_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/study_material"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authpb "github.com/Petr09Mitin/xrust-beze-back/proto/auth"
	filepb "github.com/Petr09Mitin/xrust-beze-back/proto/file"
	userpb "github.com/Petr09Mitin/xrust-beze-back/proto/user"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	httpServer *http.Server
	// grpcServer *grpc.Server
)

func main() {
	log := logger.NewLogger()
	log.Println("Starting study_material microservice...")

	cfg, err := config.NewStudyMaterial()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load study_material config")
	}

	httpPort := cfg.HTTP.Port
	// grpcPort := cfg.GRPC.Port

	db, err := initMongo(log, cfg.Mongo)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init mongo db")
	}

	// Подключение к grpc-сервисам
	userGRPCConn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", cfg.Services.UserService.Host, cfg.Services.UserService.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to user_service")
	}
	defer userGRPCConn.Close()

	fileGRPCConn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", cfg.Services.FileService.Host, cfg.Services.FileService.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to file_service")
	}
	defer fileGRPCConn.Close()

	authGRPCConn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", cfg.Services.AuthService.Host, cfg.Services.AuthService.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to auth_service")
	}
	defer authGRPCConn.Close()

	userClient := userpb.NewUserServiceClient(userGRPCConn)
	fileClient := filepb.NewFileServiceClient(fileGRPCConn)
	authClient := authpb.NewAuthServiceClient(authGRPCConn)
	
	ragClient := rag_client.NewRagClient(cfg.Services.RAGService, log)

	repo := study_material_repo.NewStudyMaterialAPIRepository(db, log)
	service := study_material_service.NewStudyMaterialAPIService(repo, userClient, fileClient, ragClient, log)

	// Каналы для сигналов завершения
	errChan := make(chan error)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// HTTP сервер
	go func() {
		router := gin.Default()
		router.Use(middleware.CORSMiddleware())
		study_material_http.NewStudyMaterialHandler(router, service, authClient)

		httpServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", httpPort),
			Handler: router,
		}
		log.Printf("HTTP server starting on port %d...", httpPort)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("failed to run HTTP server: %v", err)
		}
	}()

	// gRPC сервер
	// go func() {
	// 	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	// 	if err != nil {
	// 		errChan <- fmt.Errorf("failed to listen on port %d: %v", grpcPort, err)
	// 		return
	// 	}
	// 	grpcServer = grpc.NewServer()
	// 	studyMaterialGRPC := study_material_grpc.NewStudyMaterialServer(service, log)
	// 	pb.RegisterStudyMaterialServiceServer(grpcServer, studyMaterialGRPC)

	// 	log.Printf("gRPC server starting on port %d...", grpcPort)
	// 	if err := grpcServer.Serve(lis); err != nil {
	// 		errChan <- fmt.Errorf("failed to serve gRPC: %v", err)
	// 	}
	// }()

	// Ожидание сигнала завершения или ошибки
	select {
	case err := <-errChan:
		log.Printf("Error: %v", err)
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
	}

	log.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to shutdown http server")
	}
	// grpcServer.GracefulStop()

	log.Println("study_material microservice stopped")
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

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
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

// func initMongo(log zerolog.Logger, cfg *config.Mongo) (*mongo.Database, error) {
// 	log.Info().Msg("Connecting to MongoDB...")

// 	uri := fmt.Sprintf(
// 		"mongodb://%s:%s@%s:%d/?authSource=admin",
// 		cfg.Username,
// 		cfg.Password,
// 		cfg.Host,
// 		cfg.Port,
// 	)
// 	dbName := cfg.Database

// 	log.Info().Msgf("Connecting to MongoDB with URI: %s", uri)

// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	clientOpts := options.Client().ApplyURI(uri)

// 	client, err := mongo.Connect(clientOpts)
// 	if err != nil {
// 		log.Fatal().Err(err).Msg("failed to connect to MongoDB")
// 		return nil, err
// 	}

// 	if err := client.Ping(ctx, nil); err != nil {
// 		log.Fatal().Err(err).Msg("failed to ping MongoDB")
// 		return nil, err
// 	}

// 	log.Info().Msg("Connected to MongoDB successfully")
// 	return client.Database(dbName), nil
// }
