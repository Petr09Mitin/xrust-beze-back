package voice_recognition_repo

import (
	"context"
	chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/rs/zerolog"
)

type VoiceRecognitionPubRepo interface {
	PublishMessage(ctx context.Context, msg chat_models.Message) error
}

type VoiceRecognitionPubRepoImpl struct {
	p      message.Publisher
	topic  string
	logger zerolog.Logger
}

func NewVoiceRecognitionPubRepo(p message.Publisher, topic string, logger zerolog.Logger) VoiceRecognitionPubRepo {
	return &VoiceRecognitionPubRepoImpl{
		p:      p,
		topic:  topic,
		logger: logger,
	}
}

func (r *VoiceRecognitionPubRepoImpl) PublishMessage(_ context.Context, msg chat_models.Message) error {
	return r.p.Publish(r.topic, message.NewMessage(
		watermill.NewUUID(),
		msg.Encode(),
	))
}
