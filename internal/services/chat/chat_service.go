package chat_service

import (
	"context"
	"errors"
	"fmt"
	chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"
	message_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository"
	"time"
)

type ChatService interface {
	ProcessTextMessage(ctx context.Context, message chat_models.Message) error
	GetMessagesByChatID(ctx context.Context, chatID string) ([]chat_models.Message, error)
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
	var newMsg chat_models.Message
	var err error

	switch msg.Type {
	case chat_models.SendMessageType:
		newMsg, err = c.createTextMessage(ctx, msg)
		if err != nil {
			return err
		}
	case chat_models.UpdateMessageType:
		newMsg, err = c.updateTextMessage(ctx, msg)
		if err != nil {
			return err
		}
	case chat_models.DeleteMessageType:
		newMsg, err = c.deleteTextMessage(ctx, msg)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid msg type")
	}

	if err := c.msgRepo.PublishMessage(ctx, newMsg); err != nil {
		return fmt.Errorf("error broadcast text message: %w", err)
	}
	return nil
}

func (c *ChatServiceImpl) createTextMessage(ctx context.Context, msg chat_models.Message) (chat_models.Message, error) {
	createdAt := time.Now().Unix()
	newMsg := chat_models.Message{
		Event:     chat_models.TextMsgEvent,
		Type:      msg.Type,
		ChannelID: msg.ChannelID,
		UserID:    msg.UserID,
		Payload:   msg.Payload,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}
	newMsg, err := c.msgRepo.InsertMessage(ctx, newMsg)
	if err != nil {
		return msg, fmt.Errorf("error broadcast text message: %w", err)
	}
	fmt.Printf("new message saved: %+v\n", newMsg)
	return newMsg, nil
}

func (c *ChatServiceImpl) updateTextMessage(ctx context.Context, msg chat_models.Message) (chat_models.Message, error) {
	updatedAt := time.Now().Unix()
	newMsg := chat_models.Message{
		MessageID: msg.MessageID,
		Event:     msg.Event,
		Type:      msg.Type,
		ChannelID: msg.ChannelID,
		UserID:    msg.UserID,
		Payload:   msg.Payload,
		CreatedAt: msg.CreatedAt,
		UpdatedAt: updatedAt,
	}
	err := c.msgRepo.UpdateMessage(ctx, newMsg)
	if err != nil {
		return msg, fmt.Errorf("error broadcast text message: %w", err)
	}
	fmt.Printf("message updated: %+v\n", msg.MessageID)
	return msg, nil
}

func (c *ChatServiceImpl) deleteTextMessage(ctx context.Context, msg chat_models.Message) (chat_models.Message, error) {
	err := c.msgRepo.DeleteMessage(ctx, msg)
	if err != nil {
		return msg, fmt.Errorf("error broadcast text message: %w", err)
	}
	fmt.Printf("message deleted: %+v\n", msg.MessageID)
	return msg, nil
}

func (c *ChatServiceImpl) GetMessagesByChatID(ctx context.Context, chatID string) ([]chat_models.Message, error) {
	// TODO: add pagination
	return c.msgRepo.GetMessagesByChannelID(ctx, chatID)
}
