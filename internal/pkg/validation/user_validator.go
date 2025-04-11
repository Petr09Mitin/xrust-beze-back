package validation

import (
	"errors"
	"log"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

type ValidationErrorResponse struct {
	Error                 string   `json:"error"`
	ValidationErrorFields []string `json:"validation_error_fields"`
}

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

func BuildValidationError(err error) *ValidationErrorResponse {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		fields := make([]string, 0, len(validationErrs))
		for _, fieldErr := range validationErrs {
			fields = append(fields, fieldErr.StructField())
		}
		return &ValidationErrorResponse{
			Error:                 "validation failed: one or more fields are invalid",
			ValidationErrorFields: fields,
		}
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
