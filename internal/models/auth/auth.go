package auth

import (
	"log"
	"strings"
	"time"

	user_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/user"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	err := validate.RegisterValidation("validate-password", validatePassword)
	if err != nil {
		log.Printf("Failed to register password validation: %v", err)
	} else {
		log.Println("Successfully registered password validation")
	}
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

func (a *LoginRequest) Validate() error {
	if err := validate.Struct(a); err != nil {
		return err
	}
	return nil
}

func (r *RegisterRequest) Validate() error {
	// Валидируем встроенную структуру User
	if err := r.User.Validate(); err != nil {
		return err
	}
	// Валидируем пароль
	if err := validate.Struct(r); err != nil {
		return err
	}
	return nil
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
			log.Println("hereeee yes")
			return false // Недопустимый символ
		}
	}
	return hasUpper && hasLower && hasNumber && hasSpecial
}
