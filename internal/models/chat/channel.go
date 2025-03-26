package chat_models

import (
	"encoding/json"
	"fmt"
	user_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/user"
)

type Channel struct {
	ID          string            `json:"channel_id" bson:"_id,omitempty"`
	UserIDs     []string          `json:"user_ids" bson:"user_ids"`
	Users       []user_model.User `json:"users,omitempty" bson:"-"`
	LastMessage *Message          `json:"last_message" bson:"-"`
	Created     int64             `json:"created" bson:"created"`
	Updated     int64             `json:"updated" bson:"updated"`
}

func (c *Channel) Encode() []byte {
	result, err := json.Marshal(c)
	if err != nil {
		fmt.Println(err)
		return []byte{}
	}
	return result
}
