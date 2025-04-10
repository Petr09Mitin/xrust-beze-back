package user

import (
	"errors"
	"time"

	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/validation"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username        string             `json:"username" bson:"username" validate:"required,validate-username"`
	Email           string             `json:"email" bson:"email" validate:"required,email"`
	SkillsToLearn   []Skill            `json:"skills_to_learn" bson:"skills_to_learn"`
	SkillsToShare   []Skill            `json:"skills_to_share" bson:"skills_to_share"`
	Bio             string             `json:"bio" bson:"bio" validate:"max=1000"`
	Avatar          string             `json:"avatar" bson:"avatar"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
	LastActiveAt    time.Time          `json:"last_active_at" bson:"last_active_at"`
	PreferredFormat string             `json:"preferred_format" bson:"preferred_format" validate:"omitempty,oneof=text voice video"`
	Hrefs           []string           `json:"hrefs" bson:"hrefs"`
}

type UserToCreate struct {
	Username        string    `bson:"username"`
	Email           string    `bson:"email"`
	Password        string    `bson:"password"`
	SkillsToLearn   []Skill   `bson:"skills_to_learn"`
	SkillsToShare   []Skill   `bson:"skills_to_share"`
	Bio             string    `bson:"bio"`
	Avatar          string    `bson:"avatar"`
	PreferredFormat string    `bson:"preferred_format"`
	CreatedAt       time.Time `bson:"created_at"`
	UpdatedAt       time.Time `bson:"updated_at"`
	LastActiveAt    time.Time `bson:"last_active_at"`
	Hrefs           []string  `bson:"hrefs"`
}

type UserToUpdate struct {
	Username        string    `bson:"username"`
	Email           string    `bson:"email"`
	SkillsToLearn   []Skill   `bson:"skills_to_learn"`
	SkillsToShare   []Skill   `bson:"skills_to_share"`
	Bio             string    `bson:"bio"`
	Avatar          string    `bson:"avatar"`
	UpdatedAt       time.Time `bson:"updated_at"`
	LastActiveAt    time.Time `bson:"last_active_at"`
	PreferredFormat string    `bson:"preferred_format"`
	Hrefs           []string  `bson:"hrefs"`
}

// Навык
type Skill struct {
	Name        string `json:"name" bson:"name" validate:"required"`
	Level       string `json:"level" bson:"level" validate:"required,oneof=beginner intermediate advanced"`
	Description string `json:"description" bson:"description" validate:"max=500"`
}

// Категория + навыки (в БД)
type SkillsByCategory struct {
	Category string   `json:"category" bson:"category" validate:"required"`
	Skills   []string `json:"skills" bson:"skills"`
}

func (u *User) Validate() error {
	if err := validation.Validate.Struct(u); err != nil {
		return err
	}
	// Проверка навыков (есть хотя бы один)
	if len(u.SkillsToLearn) == 0 && len(u.SkillsToShare) == 0 {
		return errors.New("at least one skill to learn or share is required")
	}
	for _, skill := range u.SkillsToLearn {
		if err := validation.Validate.Struct(skill); err != nil {
			return errors.New("invalid skill to learn: " + err.Error())
		}
	}
	for _, skill := range u.SkillsToShare {
		if err := validation.Validate.Struct(skill); err != nil {
			return errors.New("invalid skill to share: " + err.Error())
		}
	}
	return nil
}
