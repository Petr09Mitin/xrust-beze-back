package chat_delivery

import (
	"encoding/json"
	chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"
	"strconv"
)

type MessagePresenter struct {
	MessageID string `json:"message_id"`
	Event     int    `json:"event"`
	UserID    string `json:"user_id"`
	Payload   string `json:"payload"`
	Seen      bool   `json:"seen"`
	Time      int64  `json:"time"`
}

type UserPresenter struct {
	ID   string `json:"id"`
	Name string `json:"name" binding:"required"`
}

type UserIDsPresenter struct {
	UserIDs []string `json:"user_ids"`
}

type MessagesPresenter struct {
	NextPageState string             `json:"next_ps"`
	Messages      []MessagePresenter `json:"messages"`
}

func (m *MessagePresenter) Encode() []byte {
	result, _ := json.Marshal(m)
	return result
}

func (m *MessagePresenter) ToMessage(accessToken string) (*chat_models.Message, error) {
	channelID := uint64(1)
	userID, err := strconv.ParseUint(m.UserID, 10, 64)
	if err != nil {
		return nil, err
	}
	return &chat_models.Message{
		Event:     m.Event,
		ChannelID: channelID,
		UserID:    userID,
		Payload:   m.Payload,
		Time:      m.Time,
	}, nil
}
