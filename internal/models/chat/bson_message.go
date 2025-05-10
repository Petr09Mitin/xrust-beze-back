package chat_models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type BSONMessage struct {
	MessageID       bson.ObjectID `bson:"_id,omitempty"`
	ChannelID       string        `bson:"channel_id"`
	UserID          string        `bson:"user_id"`
	PeerID          string        `bson:"peer_id"`
	Payload         string        `bson:"payload"`
	Structurized    string        `bson:"structurized,omitempty"`
	Voice           string        `bson:"voice,omitempty"`
	VoiceDuration   int64         `bson:"voice_duration,omitempty"`
	RecognizedVoice string        `bson:"recognized_voice,omitempty"`
	Attachments     []string      `bson:"attachments,omitempty"`
	CreatedAt       int64         `bson:"created_at"`
	UpdatedAt       int64         `bson:"updated_at"`
}

func (msg *BSONMessage) ToMessage() Message {
	return Message{
		MessageID:       msg.MessageID.Hex(),
		ChannelID:       msg.ChannelID,
		UserID:          msg.UserID,
		PeerID:          msg.PeerID,
		Payload:         msg.Payload,
		Structurized:    msg.Structurized,
		CreatedAt:       msg.CreatedAt,
		UpdatedAt:       msg.UpdatedAt,
		Voice:           msg.Voice,
		RecognizedVoice: msg.RecognizedVoice,
		Attachments:     msg.Attachments,
		VoiceDuration:   msg.VoiceDuration,
	}
}
