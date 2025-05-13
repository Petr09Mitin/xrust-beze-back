package voicerecognitiond

import (
	"context"
	chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"
	"github.com/Petr09Mitin/xrust-beze-back/internal/services/voice_recognition"
	"github.com/ThreeDotsLabs/watermill"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/rs/zerolog"
)

type VoiceRecognitionD struct {
	voiceRecognitionService voice_recognition.VoiceRecognitionService
	subTopicID              string
	pubTopicID              string
	router                  *message.Router
	sub                     message.Subscriber
	pub                     message.Publisher
	logger                  zerolog.Logger
}

func NewVoiceRecognitionD(
	voiceRecognitionService voice_recognition.VoiceRecognitionService,
	subTopicID string,
	pubTopicID string,
	router *message.Router,
	sub message.Subscriber,
	pub message.Publisher,
	logger zerolog.Logger,
) *VoiceRecognitionD {
	return &VoiceRecognitionD{
		voiceRecognitionService: voiceRecognitionService,
		subTopicID:              subTopicID,
		pubTopicID:              pubTopicID,
		router:                  router,
		sub:                     sub,
		pub:                     pub,
		logger:                  logger,
	}
}

func (s *VoiceRecognitionD) Run(ctx context.Context) error {
	s.registerHandler()
	return s.router.Run(ctx)
}

func (s *VoiceRecognitionD) GracefulStop() error {
	return s.router.Close()
}

func (s *VoiceRecognitionD) registerHandler() {
	s.router.AddHandler(
		"study_material_handler",
		s.subTopicID,
		s.sub,
		s.pubTopicID,
		s.pub,
		s.handleMessage,
	)
}

func (s *VoiceRecognitionD) handleMessage(msg *message.Message) ([]*message.Message, error) {
	decodedMsg, err := chat_models.DecodeToMessage(msg.Payload)
	if err != nil {
		s.logger.Err(err).Msg("failed to decode message")
		return nil, nil // need to send ack to watermill
	}
	s.logger.Info().Interface("decodedMsg", decodedMsg).Msg("decoded message")
	newMsg, err := s.voiceRecognitionService.ProcessVoiceMessage(context.Background(), decodedMsg)
	if err != nil {
		s.logger.Err(err).Msg("failed to process message")
		return nil, nil
	}
	return []*message.Message{
		message.NewMessage(
			watermill.NewUUID(),
			newMsg.Encode(),
		),
	}, nil
}
