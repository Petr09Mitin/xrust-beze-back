package chat_models

import (
	"encoding/json"
	"fmt"
)

type MsgType string
type MsgEvent string

const (
	TextMsgEvent         = MsgEvent("EventText")
	StructurizationEvent = MsgEvent("EventStructurization")
	VoiceMessageEvent    = MsgEvent("EventVoice")

	SendMessageType   = MsgType("send_message")
	UpdateMessageType = MsgType("update_message")
	DeleteMessageType = MsgType("delete_message")
)

type Message struct {
	MessageID     string         `json:"message_id,omitempty" bson:"_id,omitempty"`
	Event         MsgEvent       `json:"event,omitempty" bson:"-"`
	Type          MsgType        `json:"type,omitempty" bson:"-"`
	ChannelID     string         `json:"channel_id,omitempty" bson:"channel_id"`
	UserID        string         `json:"user_id,omitempty" bson:"user_id"`
	PeerID        string         `json:"peer_id,omitempty" bson:"peer_id"`
	ReceiverIDs   map[string]any `json:"receiver_ids,omitempty" bson:"-"`
	Payload       string         `json:"payload,omitempty" bson:"payload"`
	Structurized  string         `json:"structurized,omitempty" bson:"structurized"`
	Voice         string         `json:"voice,omitempty" bson:"voice"`
	VoiceDuration int64          `json:"voice_duration,omitempty" bson:"voice_duration"`
	Attachments   []string       `json:"attachments,omitempty" bson:"attachments"`
	CreatedAt     int64          `json:"created_at,omitempty" bson:"created_at"`
	UpdatedAt     int64          `json:"updated_at,omitempty" bson:"updated_at"`
}

func (msg *Message) Encode() []byte {
	result, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
		return []byte{}
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

func (msg *Message) SetReceiverIDs(receiverIDs []string) {
	msg.ReceiverIDs = make(map[string]any, len(receiverIDs))
	for _, receiverID := range receiverIDs {
		msg.ReceiverIDs[receiverID] = struct{}{}
	}
}
