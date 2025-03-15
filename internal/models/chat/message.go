package chat_models

type Message struct {
	MessageID uint64 `json:"message_id" bson:"message_id"`
	Event     int    `json:"event" bson:"event"`
	ChannelID uint64 `json:"channel_id" bson:"channel_id"`
	UserID    uint64 `json:"user_id" bson:"user_id"`
	Payload   string `json:"payload" bson:"payload"`
	Seen      bool   `json:"seen" bson:"seen"`
	Time      int64  `json:"time" bson:"time"`
	Structured string `json:"structured,omitempty" bson:"structured,omitempty"`
}
