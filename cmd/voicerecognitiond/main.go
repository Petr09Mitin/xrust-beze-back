package main

import (
	"context"
	"fmt"
	message_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/chat"
	voice_recognition_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/voice_recognition"
	"github.com/Petr09Mitin/xrust-beze-back/internal/router/daemons/voicerecognitiond"
	"github.com/Petr09Mitin/xrust-beze-back/internal/services/voice_recognition"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
	infrakafka "github.com/Petr09Mitin/xrust-beze-back/internal/pkg/kafka"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/logger"
)

func main() {
	// init logger
	log := logger.NewLogger()
	log.Println("Starting voicerecognitiond...")

	// init cfg
	cfg, err := config.NewVoiceRecognitionD()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load voicerecognitiond config")
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
	messagesCollection := client.Database(cfg.Mongo.Database).Collection("messages")
	messagesRepo := message_repo.NewMessageRepo(messagesCollection, log)
	aiVoiceRecognitionRepo := voice_recognition_repo.NewVoiceRecognitionRepo(cfg.Services.AIVoiceRecognition, log)

	// init kafka sub
	kafkaSub, err := infrakafka.NewKafkaSubscriber(cfg.Kafka)
	if err != nil {
		log.Err(err).Msg("failed to connect to kafka sub")
		return
	}
	// init kafka sub
	kafkaPub, err := infrakafka.NewKafkaPublisher(cfg.Kafka)
	if err != nil {
		log.Err(err).Msg("failed to connect to kafka pub")
		return
	}
	brokerRouter, err := infrakafka.NewBrokerRouter()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize kafka msg_router")
		return
	}
	voiceRecognitionService := voice_recognition.NewVoiceRecognitionService(aiVoiceRecognitionRepo, messagesRepo, log)
	d := voicerecognitiond.NewVoiceRecognitionD(
		voiceRecognitionService,
		cfg.Kafka.VoiceRecognitionNewVoiceTopic,
		cfg.Kafka.VoiceRecognitionVoiceProcessedTopic,
		brokerRouter,
		kafkaSub,
		kafkaPub,
		log,
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		err := d.GracefulStop()
		if err != nil {
			log.Error().Err(err).Msg("failed to gracefully stop voicerecognitiond")
		}
	}()
	if err = d.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("error running voicerecognitiond")
		return
	}
}
