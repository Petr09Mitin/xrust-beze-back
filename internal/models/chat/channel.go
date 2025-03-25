package chat_models

import (
	"encoding/json"
	"fmt"
)

type Channel struct {
	ID          string   `json:"channel_id" bson:"_id,omitempty"`
	UserIDs     []string `json:"user_ids" bson:"user_ids"`
	LastMessage *Message `json:"last_message" bson:"-"`
	Created     int64    `json:"created" bson:"created"`
	Updated     int64    `json:"updated" bson:"updated"`
}

func (c *Channel) Encode() []byte {
	result, err := json.Marshal(c)
	if err != nil {
		fmt.Println(err)
		return []byte{}
	}
	return result
}
