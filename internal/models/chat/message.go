package chat_models

type Message struct {
	MessageID uint64 `json:"message_id"`
	Event     int    `json:"event"`
	ChannelID uint64 `json:"channel_id"`
	UserID    uint64 `json:"user_id"`
	Payload   string `json:"payload"`
	Seen      bool   `json:"seen"`
	Time      int64  `json:"time"`
}
