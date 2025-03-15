package main

import (
	"github.com/Petr09Mitin/xrust-beze-back/internal/router"
	chat_service "github.com/Petr09Mitin/xrust-beze-back/internal/services/chat"
)

func main() {
	chatService := chat_service.NewChatService()
	c := router.NewChat(chatService)
	err := c.Start()
	if err != nil {
		panic(err)
	}
}
