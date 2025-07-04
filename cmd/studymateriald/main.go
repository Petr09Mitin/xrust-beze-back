package main

import (
	"context"
	"fmt"
	"github.com/Petr09Mitin/xrust-beze-back/internal/repository/rag_client"

	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
	infrakafka "github.com/Petr09Mitin/xrust-beze-back/internal/pkg/kafka"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/logger"
	study_material_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/study_material"
	study_materiald "github.com/Petr09Mitin/xrust-beze-back/internal/router/daemons/study_material"
	study_material_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/study_material"
	filepb "github.com/Petr09Mitin/xrust-beze-back/proto/file"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// init logger
	log := logger.NewLogger()
	log.Println("Starting studymateriald...")

	// init cfg
	cfg, err := config.NewStudyMaterial()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load studymateriald config")
	}

	// init mongo
	client, err := mongo.Connect(options.Client().ApplyURI(fmt.Sprintf(
		"mongodb://%s:%s@%s:%d",
		cfg.Mongo.Username,
		cfg.Mongo.Password,
		cfg.Mongo.Host,
		cfg.Mongo.Port,
	)))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to mongodb")
		return
	}
	studyMaterialCollection := client.Database(cfg.Mongo.Database).Collection("study_materials")
	studyMaterialRepo := study_material_repo.NewStudyMaterialRepo(studyMaterialCollection, log)

	// init kafka sub
	kafkaSub, err := infrakafka.NewKafkaSubscriber(cfg.Kafka)
	if err != nil {
		log.Err(err).Msg("failed to connect to kafka sub")
		return
	}
	brokerRouter, err := infrakafka.NewBrokerRouter()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize kafka msg_router")
		return
	}

	// init file service client
	fileGRPCConn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", cfg.Services.FileService.Host, cfg.Services.FileService.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to file_service")
		return
	}
	fileGRPCClient := filepb.NewFileServiceClient(fileGRPCConn)
	fileServiceClient := study_material_repo.NewFileRepo(fileGRPCClient, log)

	// init ML repo
	mlTaggerRepo := study_material_repo.NewMLTaggerRepo(cfg.Services.MLTags, log)

	// init RAG repo
	RAGRepo := rag_client.NewRagClient(cfg.Services.RAGService, log)

	studyMaterialService := study_material_service.NewStudyMaterialService(studyMaterialRepo, mlTaggerRepo, fileServiceClient, RAGRepo, log)
	d := study_materiald.NewStudyMaterialD(studyMaterialService, cfg.Kafka.StudyMaterialTopic, brokerRouter, kafkaSub, log)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		err := d.GracefulStop()
		if err != nil {
			log.Error().Err(err).Msg("failed to gracefully stop studymateriald")
		}
	}()
	if err = d.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("error running studymateriald")
		return
	}
}
