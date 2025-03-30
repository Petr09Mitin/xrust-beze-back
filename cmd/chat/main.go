package main

import (
	infrakafka "github.com/Petr09Mitin/xrust-beze-back/internal/pkg/kafka"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/logger"
	channelrepo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/channel"
	message_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/chat"
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
	kafkaPub, err := infrakafka.NewKafkaPublisher()
	if err != nil {
		log.Err(err)
		return
	}
	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://admin:admin@mongo_db:27017"))
	if err != nil {
		log.Err(err)
		return
	}
	msgsCollection := client.Database("xrust_beze").Collection("messages")
	chanCollection := client.Database("xrust_beze").Collection("channels")
	msgRepo := message_repo.NewMessageRepo(kafkaPub, msgsCollection, log)
	chanRepo := channelrepo.NewChannelRepository(chanCollection, log)
	userGRPCConn, err := grpc.NewClient("user_service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Err(err)
		return
	}
	userGRPCClient := pb.NewUserServiceClient(userGRPCConn)
	chatService := chat_service.NewChatService(msgRepo, chanRepo, userGRPCClient, log)
	m := melody.New()
	kafkaSub, err := infrakafka.NewKafkaSubscriber()
	if err != nil {
		log.Err(err)
		return
	}
	msgRouter, err := infrakafka.NewBrokerRouter()
	msgSub, err := chat.NewMessageSubscriber(msgRouter, kafkaSub, m, log)
	if err != nil {
		log.Err(err)
		return
	}
	c, err := chat.NewChat(chatService, msgSub, m, log)
	if err != nil {
		log.Err(err)
		return
	}
	err = c.Start()
	if err != nil {
		c.Stop()
		log.Err(err)
		return
	}
}
