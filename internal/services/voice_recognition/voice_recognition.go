package voice_recognition

import (
	"context"
	"errors"
	chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"
	message_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/chat"
	voice_recognition_repo "github.com/Petr09Mitin/xrust-beze-back/internal/repository/voice_recognition"
	"github.com/rs/zerolog"
	"time"
)

type VoiceRecognitionService interface {
	ProcessVoiceMessage(ctx context.Context, msg *chat_models.Message) (*chat_models.Message, error)
}

type VoiceRecognitionServiceImpl struct {
	voiceRecognitionRepo voice_recognition_repo.VoiceRecognitionRepo
	messagesRepo         message_repo.MessageRepo
	logger               zerolog.Logger
}

func NewVoiceRecognitionService(voiceRecognitionRepo voice_recognition_repo.VoiceRecognitionRepo, messagesRepo message_repo.MessageRepo, logger zerolog.Logger) VoiceRecognitionService {
	return &VoiceRecognitionServiceImpl{
		voiceRecognitionRepo: voiceRecognitionRepo,
		messagesRepo:         messagesRepo,
		logger:               logger,
	}
}

func (v *VoiceRecognitionServiceImpl) ProcessVoiceMessage(ctx context.Context, msg *chat_models.Message) (*chat_models.Message, error) {
	if msg.Voice == "" {
		v.logger.Error().Msg("voice is empty")
		return nil, errors.New("voice is empty")
	}
	recognizedVoice, err := v.voiceRecognitionRepo.SendVoiceRecognitionRequest(ctx, msg.Voice)
	if err != nil {
		v.logger.Error().Err(err).Msg("failed to send voice recognition request")
		return nil, err
	}
	msg.RecognizedVoice = recognizedVoice
	msg.UpdatedAt = time.Now().Unix()
	err = v.messagesRepo.UpdateMessage(ctx, *msg)
	if err != nil {
		v.logger.Error().Err(err).Msg("failed to update msg")
		return nil, err
	}
	msg.Event = chat_models.VoiceRecognizedEvent
	return msg, nil
}
