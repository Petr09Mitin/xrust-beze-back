package main

import (
	infrakafka "github.com/Petr09Mitin/xrust-beze-back/internal/pkg/kafka"
	message_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository"
	"github.com/Petr09Mitin/xrust-beze-back/internal/router"
	"github.com/Petr09Mitin/xrust-beze-back/internal/router/chat"
	chat_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/chat"
	"github.com/olahol/melody"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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
	collection := client.Database("xrust_beze").Collection("chats")
	msgRepo := message_repo.NewMessageRepo(kafkaPub, collection)
	chatService := chat_service.NewChatService(msgRepo)
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
	c := router.NewChat(chatService, msgSub, m)

	err = c.Start()
	if err != nil {
		c.Stop()
		log.Fatal(err)
	}
}
