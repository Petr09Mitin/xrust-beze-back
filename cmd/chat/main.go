package main

import (
	infrakafka "github.com/Petr09Mitin/xrust-beze-back/internal/pkg/kafka"
	message_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository"
	"github.com/Petr09Mitin/xrust-beze-back/internal/router"
	"github.com/Petr09Mitin/xrust-beze-back/internal/router/chat"
	chat_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/chat"
	"github.com/olahol/melody"
	"log"
)

func main() {
	kafkaPub, err := infrakafka.NewKafkaPublisher()
	if err != nil {
		log.Fatal(err)
	}
	msgRepo := message_repo.NewMessageRepo(kafkaPub)
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
		log.Fatal(err)
	}
}
