package message_repo

import (
	"context"
	chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

const (
	MessagePubTopic = "xb.msg.pub"
)

type MessageRepo interface {
	InsertMessage(ctx context.Context, msg chat_models.Message) (chat_models.Message, error)
	PublishMessage(ctx context.Context, msg chat_models.Message) error
}

type MessageRepoImpl struct {
	p message.Publisher
}

func NewMessageRepo(p message.Publisher) MessageRepo {
	return &MessageRepoImpl{
		p: p,
	}
}

func (m *MessageRepoImpl) InsertMessage(_ context.Context, msg chat_models.Message) (chat_models.Message, error) {
	return chat_models.Message{
		Payload:   "xdd",
		ChannelID: "1",
		UserID:    "2",
	}, nil
}

func (m *MessageRepoImpl) PublishMessage(_ context.Context, msg chat_models.Message) error {
	return m.p.Publish(MessagePubTopic, message.NewMessage(
		watermill.NewUUID(),
		msg.Encode(),
	))
}
