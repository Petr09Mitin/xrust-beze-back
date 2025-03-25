package chat_models

import "go.mongodb.org/mongo-driver/v2/bson"

type BSONMessage struct {
	MessageID bson.ObjectID `bson:"_id,omitempty"`
	ChannelID string        `bson:"channel_id"`
	UserID    string        `bson:"user_id"`
	PeerID    string        `bson:"peer_id"`
	Payload   string        `bson:"payload"`
	CreatedAt int64         `bson:"created_at"`
	UpdatedAt int64         `bson:"updated_at"`
}

func (msg *BSONMessage) ToMessage() Message {
	return Message{
		MessageID: msg.MessageID.Hex(),
		ChannelID: msg.ChannelID,
		UserID:    msg.UserID,
		PeerID:    msg.PeerID,
		Payload:   msg.Payload,
		CreatedAt: msg.CreatedAt,
		UpdatedAt: msg.UpdatedAt,
	}
}
