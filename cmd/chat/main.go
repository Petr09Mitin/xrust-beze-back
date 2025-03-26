package main

import (
	infrakafka "github.com/Petr09Mitin/xrust-beze-back/internal/pkg/kafka"
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
	"log"
)

func main() {
	kafkaPub, err := infrakafka.NewKafkaPublisher()
	if err != nil {
		log.Fatal(err)
	}
	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://admin:admin@mongo_db:27017"))
	if err != nil {
		log.Fatal(err)
	}
	msgsCollection := client.Database("xrust_beze").Collection("messages")
	chanCollection := client.Database("xrust_beze").Collection("channels")
	msgRepo := message_repo.NewMessageRepo(kafkaPub, msgsCollection)
	chanRepo := channelrepo.NewChannelRepository(chanCollection)
	userGRPCConn, err := grpc.NewClient("user_service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	userGRPCClient := pb.NewUserServiceClient(userGRPCConn)
	chatService := chat_service.NewChatService(msgRepo, chanRepo, userGRPCClient)
	m := melody.New()
	kafkaSub, err := infrakafka.NewKafkaSubscriber()
	if err != nil {
		log.Fatal(err)
	}
	msgRouter, err := infrakafka.NewBrokerRouter()
	msgSub, err := chat.NewMessageSubscriber(msgRouter, kafkaSub, m)
	if err != nil {
		log.Fatal(err)
	}
	c := chat.NewChat(chatService, msgSub, m)

	err = c.Start()
	if err != nil {
		c.Stop()
		log.Fatal(err)
	}
}
