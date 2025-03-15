package chat_service

import chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"

type ChatService interface {
	ProcessTextMessage(message chat_models.Message) error
}

type ChatServiceImpl struct {
}

func NewChatService() ChatService {
	return &ChatServiceImpl{}
}

func (c *ChatServiceImpl) ProcessTextMessage(message chat_models.Message) error {
	
	return nil
}
