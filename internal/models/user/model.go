package user

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	validate *validator.Validate
	// emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

func init() {
	validate = validator.New()
}

// Модель пользователя
type User struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username        string             `json:"username" bson:"username" validate:"required,min=3,max=50"`
	Email           string             `json:"email" bson:"email" validate:"required,email"`
	SkillsToLearn   []Skill            `json:"skills_to_learn" bson:"skills_to_learn"`
	SkillsToShare   []Skill            `json:"skills_to_share" bson:"skills_to_share"`
	Bio             string             `json:"bio" bson:"bio" validate:"max=1000"`
	AvatarURL       string             `json:"avatar_url" bson:"avatar_url"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
	LastActiveAt    time.Time          `json:"last_active_at" bson:"last_active_at"`
	PreferredFormat string             `json:"preferred_format" bson:"preferred_format" validate:"omitempty,oneof=text voice video"`
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

// Проверяет валидность модели пользователя
func (u *User) Validate() error {
	if err := validate.Struct(u); err != nil {
		return err
	}

	// if !emailRegex.MatchString(u.Email) {
	// 	return errors.New("invalid email format")
	// }

	// Проверка навыков (есть хотя бы один)
	if len(u.SkillsToLearn) == 0 && len(u.SkillsToShare) == 0 {
		return errors.New("at least one skill to learn or share is required")
	}

	// Валидация каждого навыка
	for _, skill := range u.SkillsToLearn {
		if err := validate.Struct(skill); err != nil {
			return errors.New("invalid skill to learn: " + err.Error())
		}
	}

	for _, skill := range u.SkillsToShare {
		if err := validate.Struct(skill); err != nil {
			return errors.New("invalid skill to share: " + err.Error())
		}
	}

	return nil
}
