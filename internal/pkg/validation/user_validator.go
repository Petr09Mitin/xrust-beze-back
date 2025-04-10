package validation

import (
	"log"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func Init() error {
	Validate = validator.New()

	if err := Validate.RegisterValidation("validate-username", ValidateUsername); err != nil {
		return err
	}
	if err := Validate.RegisterValidation("validate-password", ValidatePassword); err != nil {
		return err
	}
	return nil
}

func ValidateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	if len(username) < 3 || len(username) > 50 {
		return false
	}
	if strings.HasPrefix(username, ".") || strings.HasSuffix(username, ".") {
		return false
	}
	if strings.Contains(username, "..") {
		return false
	}
	for _, c := range username {
		switch {
		case ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z'):
			continue
		case unicode.IsDigit(c):
			continue
		case c == '.' || c == '_':
			continue
		default:
			return false
		}
	}
	return true
}

func ValidatePassword(fl validator.FieldLevel) bool {
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
