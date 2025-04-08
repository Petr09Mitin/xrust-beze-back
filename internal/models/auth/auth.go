package auth

import (
	"strings"
	"time"

	user_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/user"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	_ = validate.RegisterValidation("validate-password", validatePassword)
}

type RegisterRequest struct {
	user_model.User
	Password string `json:"password" bson:"password" validate:"required,validate-password"`
}

type LoginRequest struct {
	Email    string `json:"email" bson:"email" validate:"required,email"`
	Password string `json:"password" bson:"password" validate:"required"`
}

type Session struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	UserID    string    `json:"user_id" bson:"user_id"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	ExpiresAt time.Time `json:"expires_at" bson:"expires_at"`
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 || len(password) > 64 {
		return false
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, c := range password {
		switch {
		case 'A' <= c && c <= 'Z':
			hasUpper = true
		case 'a' <= c && c <= 'z':
			hasLower = true
		case '0' <= c && c <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*", c):
			hasSpecial = true
		default:
			return false // Недопустимый символ
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

func (a *LoginRequest) Validate() error {
	if err := validate.Struct(a); err != nil {
		return err
	}
	return nil
}

// Проверяет валидность запроса регистрации
func (r *RegisterRequest) Validate() error {
	if err := validate.Struct(r); err != nil {
		return err
	}

	// Дополнительная валидация пользователя
	if err := r.User.Validate(); err != nil {
		return err
	}

	return nil
}
