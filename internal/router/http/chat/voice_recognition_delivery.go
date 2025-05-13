package chat

import (
	"context"
	chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/olahol/melody"
	"github.com/rs/zerolog"
)

type VoiceRecognitionSubscriber struct {
	subTopic string
	router   *message.Router
	sub      message.Subscriber
	m        *melody.Melody
	logger   zerolog.Logger
}

func NewVoiceRecognitionSubscriber(subTopic string, router *message.Router, sub message.Subscriber, m *melody.Melody, logger zerolog.Logger) (*VoiceRecognitionSubscriber, error) {
	return &VoiceRecognitionSubscriber{
		subTopic: subTopic,
		router:   router,
		sub:      sub,
		m:        m,
		logger:   logger,
	}, nil
}

func (s *VoiceRecognitionSubscriber) HandleMessage(msg *message.Message) error {
	decodedMsg, err := chat_models.DecodeToMessage(msg.Payload)
	if err != nil {
		s.logger.Err(err).Msg("voice recognition failed to decode message")
		return err
	}
	s.logger.Printf("voice recognition sub got the message: %+v\n", decodedMsg)
	err = s.sendMessage(context.Background(), *decodedMsg)
	if err != nil {
		s.logger.Err(err).Msg("voice recognition failed to send message")
		return err
	}
	return nil
}

func (s *VoiceRecognitionSubscriber) RegisterHandler() {
	s.router.AddNoPublisherHandler(
		"voice_recognition_voice_processed_handler1",
		s.subTopic,
		s.sub,
		s.HandleMessage,
	)
}

func (s *VoiceRecognitionSubscriber) Run() error {
	return s.router.Run(context.Background())
}

func (s *VoiceRecognitionSubscriber) GracefulStop() error {
	return s.router.Close()
}

func (s *VoiceRecognitionSubscriber) sendMessage(_ context.Context, message chat_models.Message) error {
	messageWithoutReceivers := message
	messageWithoutReceivers.ReceiverIDs = nil
	return s.m.BroadcastFilter(messageWithoutReceivers.Encode(), func(sess *melody.Session) bool {
		userIDData, exist := sess.Get(UserIDSessionParam)
		if !exist {
			return false
		}
		userID, ok := userIDData.(string)
		if !ok {
			return false
		}
		_, ok = message.ReceiverIDs[userID]
		return ok
	})
}
