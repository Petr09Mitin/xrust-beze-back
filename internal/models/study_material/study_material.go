package study_material_models

import (
	chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"
	user_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/user"
)

type StudyMaterial struct {
	ID       string          `json:"id" bson:"_id"`
	Name     string          `json:"name" bson:"name"`
	Filename string          `json:"filename" bson:"filename"`
	Tags     []string        `json:"tags" bson:"tags"`
	AuthorID string          `json:"author_id" bson:"author_id"`
	Author   user_model.User `json:"author" bson:"-"`
	Created  int64           `json:"created" bson:"created"`
	Updated  int64           `json:"updated" bson:"updated"`
}

type AttachmentToParse struct {
	Filename         string                `json:"filename"`
	AuthorID         string                `json:"author_id"`
	CurrMessageText  string                `json:"curr_message_text"`
	PrevMessageTexts []chat_models.Message `json:"prev_messages_texts,omitempty"`
}
