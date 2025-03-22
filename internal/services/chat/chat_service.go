package chat_service

import (
	"context"
	"fmt"
	chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"
	message_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository"
)

type ChatService interface {
	ProcessTextMessage(ctx context.Context, message chat_models.Message) error
}

type ChatServiceImpl struct {
	msgRepo message_repo.MessageRepo
}

func NewChatService(msgRepo message_repo.MessageRepo) ChatService {
	return &ChatServiceImpl{
		msgRepo: msgRepo,
	}
}

func (c *ChatServiceImpl) ProcessTextMessage(ctx context.Context, msg chat_models.Message) error {
	newMsg, err := c.msgRepo.InsertMessage(ctx, msg)
	if err != nil {
		return fmt.Errorf("error broadcast text message: %w", err)
	}
	fmt.Printf("new message saved: %+v\n", newMsg)

	if err := c.msgRepo.PublishMessage(ctx, newMsg); err != nil {
		return fmt.Errorf("error broadcast text message: %w", err)
	}
	return nil
}
