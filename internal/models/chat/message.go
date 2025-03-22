package chat_models

import (
	"encoding/json"
	"fmt"
)

const (
	SendMessageType   = MsgType("send_message")
	UpdateMessageType = MsgType("update_message")
	DeleteMessageType = MsgType("delete_message")
)

type MsgType string

type Message struct {
	MessageID uint64  `json:"message_id,omitempty"`
	Event     int     `json:"event,omitempty"`
	Type      MsgType `json:"type,omitempty"`
	ChannelID string  `json:"channel_id,omitempty"`
	UserID    string  `json:"user_id,omitempty"`
	Payload   string  `json:"payload,omitempty"`
	CreatedAt int64   `json:"created_at,omitempty"`
	UpdatedAt int64   `json:"updated_at,omitempty"`
}

func (msg *Message) Encode() []byte {
	result, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
	}
	return result
}

func DecodeToMessage(msg []byte) (*Message, error) {
	var message Message
	err := json.Unmarshal(msg, &message)
	if err != nil {
		return nil, err
	}

	return &message, nil
}
