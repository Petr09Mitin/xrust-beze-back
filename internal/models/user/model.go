package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User представляет модель пользователя
type User struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username        string             `json:"username" bson:"username"`
	Email           string             `json:"email" bson:"email"`
	SkillsToLearn   []Skill            `json:"skills_to_learn" bson:"skills_to_learn"`
	SkillsToShare   []Skill            `json:"skills_to_share" bson:"skills_to_share"`
	Bio             string             `json:"bio" bson:"bio"`
	AvatarURL       string             `json:"avatar_url" bson:"avatar_url"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
	LastActiveAt    time.Time          `json:"last_active_at" bson:"last_active_at"`
	PreferredFormat string             `json:"preferred_format" bson:"preferred_format"` // "text", "voice", "video"
}

// Skill представляет навык пользователя
type Skill struct {
	Name        string `json:"name" bson:"name"`
	Level       string `json:"level" bson:"level"` // "beginner", "intermediate", "advanced"
	Description string `json:"description" bson:"description"`
} 