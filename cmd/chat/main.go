package main

import (
	"fmt"

	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
	infrakafka "github.com/Petr09Mitin/xrust-beze-back/internal/pkg/kafka"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/logger"
	channelrepo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/channel"
	message_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/chat"
	structurization_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/structurization"
	"github.com/Petr09Mitin/xrust-beze-back/internal/router/http/chat"
	chat_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/chat"
	pb "github.com/Petr09Mitin/xrust-beze-back/proto/user"
	"github.com/olahol/melody"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log := logger.NewLogger()
	cfg, err := config.NewChat()
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("failed to load config: %v", err))
		return
	}
	kafkaPub, err := infrakafka.NewKafkaPublisher(cfg.Kafka)
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("failed to create kafka publisher: %v", err))
		return
	}
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
	msgsCollection := client.Database(cfg.Mongo.Database).Collection("messages")
	chanCollection := client.Database(cfg.Mongo.Database).Collection("channels")
	msgRepo := message_repo.NewMessageRepo(kafkaPub, msgsCollection, log)
	chanRepo := channelrepo.NewChannelRepository(chanCollection, log)
	userGRPCConn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", cfg.Services.UserService.Host, cfg.Services.UserService.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to user_service")
		return
	}
	userGRPCClient := pb.NewUserServiceClient(userGRPCConn)
	structurizationRepo := structurization_repo.NewStructurizationRepository(cfg.Services.StructurizationService, log)
	chatService := chat_service.NewChatService(msgRepo, chanRepo, structurizationRepo, userGRPCClient, log, cfg)
	m := melody.New()
	m.Config.MaxMessageSize = 1 << 20
	kafkaSub, err := infrakafka.NewKafkaSubscriber(cfg.Kafka)
	if err != nil {
		log.Err(err).Msg("failed to connect to kafka sub")
		return
	}
	msgRouter, err := infrakafka.NewBrokerRouter()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize kafka msg_router")
		return
	}
	msgSub, err := chat.NewMessageSubscriber(msgRouter, kafkaSub, m, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to kafka msg_sub")
		return
	}
	c, err := chat.NewChat(chatService, msgSub, m, log, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create chat")
		return
	}
	err = c.Start()
	if err != nil {
		c.Stop()
		log.Fatal().Err(err).Msg("failed to start chat")
		return
	}
}
