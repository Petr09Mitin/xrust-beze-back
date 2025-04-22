package study_material_models

import (
	"encoding/json"

	user_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/user"
)

type StudyMaterial struct {
	ID       string          `json:"id" bson:"_id,omitempty"`
	Name     string          `json:"name" bson:"name"`
	Filename string          `json:"filename" bson:"filename"`
	Tags     []string        `json:"tags" bson:"tags"`
	AuthorID string          `json:"author_id" bson:"author_id"`
	Author   user_model.User `json:"author" bson:"-"`
	Created  int64           `json:"created" bson:"created"`
	Updated  int64           `json:"updated" bson:"updated"`
}

// ParsedAttachmentResponse is a type that represents response from ML service
type ParsedAttachmentResponse struct {
	IsStudyMaterial bool           `json:"is_study_material"`
	StudyMaterial   *StudyMaterial `json:"study_material,omitempty"`
}

type AttachmentToParse struct {
	Filename         string   `json:"filename"`
	AuthorID         string   `json:"author_id"`
	CurrMessageText  string   `json:"curr_message_text"`
	PrevMessageTexts []string `json:"prev_messages_texts"`
}

type AttachmentToParseRequest struct {
	Filename string `json:"file_id"`
	S3Bucket string `json:"bucket_name"`
}

func (a *AttachmentToParse) Encode() []byte {
	data, _ := json.Marshal(a)
	return data
}

func DecodeToAttachmentToParse(data []byte) (*AttachmentToParse, error) {
	result := &AttachmentToParse{}
	err := json.Unmarshal(data, result)
	return result, err
}
