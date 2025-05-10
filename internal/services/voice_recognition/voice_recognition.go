package voice_recognition

import (
	"context"
	"errors"
	"fmt"
	chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"
	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
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
	cfg                  *config.VoiceRecognitionD
}

func NewVoiceRecognitionService(voiceRecognitionRepo voice_recognition_repo.VoiceRecognitionRepo, messagesRepo message_repo.MessageRepo, logger zerolog.Logger, cfg *config.VoiceRecognitionD) VoiceRecognitionService {
	return &VoiceRecognitionServiceImpl{
		voiceRecognitionRepo: voiceRecognitionRepo,
		messagesRepo:         messagesRepo,
		logger:               logger,
		cfg:                  cfg,
	}
}

func (v *VoiceRecognitionServiceImpl) ProcessVoiceMessage(ctx context.Context, msg *chat_models.Message) (*chat_models.Message, error) {
	if msg.Voice == "" {
		v.logger.Error().Msg("voice is empty")
		return nil, errors.New("voice is empty")
	}
	recognizedVoice, err := v.trySendVoiceRecognitionRequest(ctx, msg.Voice)
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

func (v *VoiceRecognitionServiceImpl) trySendVoiceRecognitionRequest(ctx context.Context, filename string) (string, error) {
	newCtx, cancel := context.WithTimeout(
		ctx,
		time.Duration(v.cfg.Services.AIVoiceRecognition.Timeout)*time.Second,
	)
	defer cancel()
	i := v.cfg.Services.AIVoiceRecognition.MaxRetries
loop:
	for i > 0 {
		select {
		case <-newCtx.Done():
			return "", custom_errors.ErrRequestTimeout
		default:
			i--
			recognized, err := v.voiceRecognitionRepo.SendVoiceRecognitionRequest(newCtx, filename)
			if err != nil {
				v.logger.Error().Err(err).Msg(fmt.Sprintf("trySendVoiceRecognitionRequest failed, %d retries remaining", i))
				continue loop
			}
			return recognized, nil
		}
	}

	return "", custom_errors.ErrMaxRetriesExceeded
}
