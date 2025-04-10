package auth

import (
	"time"

	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	user_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/user"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/validation"
)

type RegisterRequest struct {
	user_model.User
	Password string `json:"password" bson:"password" validate:"required,validate-password"`
}

type LoginRequest struct {
	Email    string `json:"email" bson:"email" validate:"omitempty,email"`
	Username string `json:"username" bson:"username" validate:"omitempty"`
	Password string `json:"password" bson:"password" validate:"required"`
}

type Session struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	UserID    string    `json:"user_id" bson:"user_id"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	ExpiresAt time.Time `json:"expires_at" bson:"expires_at"`
}

func (a *LoginRequest) Validate() error {
	if err := validation.Validate.Struct(a); err != nil {
		return err
	}
	if a.Email == "" && a.Username == "" {
		return custom_errors.ErrMissingLoginField
	}
	if a.Email != "" && a.Username != "" {
		return custom_errors.ErrTooManyLoginFields
	}
	return nil
}

func (r *RegisterRequest) Validate() error {
	// Валидируем встроенную структуру User
	if err := r.User.Validate(); err != nil {
		return err
	}
	// Валидируем пароль
	if err := validation.Validate.Struct(r); err != nil {
		return err
	}
	return nil
}
