package chat

import (
	"context"
	chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/olahol/melody"
	"github.com/rs/zerolog"
)

const (
	UserIDSessionParam = "user_id_session"
)

type MessageSubscriber struct {
	subscriberID string
	router       *message.Router
	sub          message.Subscriber
	m            *melody.Melody
	logger       zerolog.Logger
}

func NewMessageSubscriber(router *message.Router, sub message.Subscriber, m *melody.Melody, logger zerolog.Logger) (*MessageSubscriber, error) {
	return &MessageSubscriber{
		subscriberID: "xb.msg.pub",
		router:       router,
		sub:          sub,
		m:            m,
		logger:       logger,
	}, nil
}

func (s *MessageSubscriber) HandleMessage(msg *message.Message) error {
	decodedMsg, err := chat_models.DecodeToMessage(msg.Payload)
	if err != nil {
		s.logger.Err(err).Msg("failed to decode message")
		return err
	}
	s.logger.Printf("sub got the message: %+v\n", decodedMsg)
	return s.sendMessage(context.Background(), *decodedMsg)
}

func (s *MessageSubscriber) RegisterHandler() {
	s.router.AddNoPublisherHandler(
		"message_handler",
		s.subscriberID,
		s.sub,
		s.HandleMessage,
	)
}

func (s *MessageSubscriber) Run() error {
	return s.router.Run(context.Background())
}

func (s *MessageSubscriber) GracefulStop() error {
	return s.router.Close()
}

func (s *MessageSubscriber) sendMessage(_ context.Context, message chat_models.Message) error {
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
