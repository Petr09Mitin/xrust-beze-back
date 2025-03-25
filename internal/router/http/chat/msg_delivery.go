package chat

import (
	"context"
	"fmt"
	chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/olahol/melody"
)

const (
	ChannelIDSessionParam = "channel_id_session"
	UserIDSessionParam    = "user_id_session"
)

type MessageSubscriber struct {
	subscriberID string
	router       *message.Router
	sub          message.Subscriber
	m            *melody.Melody
}

func NewMessageSubscriber(router *message.Router, sub message.Subscriber, m *melody.Melody) (*MessageSubscriber, error) {
	return &MessageSubscriber{
		subscriberID: "xb.msg.pub",
		router:       router,
		sub:          sub,
		m:            m,
	}, nil
}

func (s *MessageSubscriber) HandleMessage(msg *message.Message) error {
	decodedMsg, err := chat_models.DecodeToMessage(msg.Payload)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("sub got the msg", msg)
	return s.sendMessage(context.Background(), decodedMsg)
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

func (s *MessageSubscriber) sendMessage(_ context.Context, message *chat_models.Message) error {
	return s.m.BroadcastFilter(message.Encode(), func(sess *melody.Session) bool {
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
