package message_repo

import (
	"context"
	chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/rs/zerolog"
)

const (
	MessagePubTopic = "xb.msg.pub"
)

type MessagePubRepo interface {
	PublishMessage(ctx context.Context, msg chat_models.Message) error
}

type MessagePubRepoImpl struct {
	p      message.Publisher
	logger zerolog.Logger
}

func NewMessagePubRepo(p message.Publisher, logger zerolog.Logger) MessagePubRepo {
	return &MessagePubRepoImpl{
		p:      p,
		logger: logger,
	}
}

func (m *MessagePubRepoImpl) PublishMessage(_ context.Context, msg chat_models.Message) error {
	return m.p.Publish(MessagePubTopic, message.NewMessage(
		watermill.NewUUID(),
		msg.Encode(),
	))
}
