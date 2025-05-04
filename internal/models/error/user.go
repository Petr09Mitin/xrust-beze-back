package custom_errors

import (
	"errors"
	"fmt"
)

// type ProfanityError struct {
// 	FieldName string
// }

// func (e *ProfanityError) Error() string {
// 	return fmt.Sprintf("profanity detected in %s", e.FieldName)
// }

type ProfanityAggregateError struct {
	Fields []string `json:"profanity_error_fields"`
}

func (e *ProfanityAggregateError) Error() string {
	return "profanity detected"
}

func (e *ProfanityAggregateError) IsEmpty() bool {
	return len(e.Fields) == 0
}

func (e *ProfanityAggregateError) Add(field string) {
	e.Fields = append(e.Fields, field)
}

var (
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrUserNotExists         = errors.New("user does not exist")
	ErrNoUsername            = fmt.Errorf("%w: no username provided", ErrBadRequest)
	ErrMissingPassword       = errors.New("password field missing or not a string")
	ErrValidationFailed      = errors.New("validation failed: one or more fields are invalid")
	ErrModerationUnavailable = errors.New("ml moderation service unavailable")
	// ErrUsernameProfanityDetected = errors.New("profanity detected in username")
	// ErrBioProfanityDetected      = errors.New("profanity detected in bio")
	// ErrProfanityDetected = &ProfanityError{FieldName: ""}
	ErrDuplicateReview  = fmt.Errorf("%w: duplicate review", ErrBadRequest)
	ErrCanNotSelfReview = fmt.Errorf("%w: can not create self-review", ErrBadRequest)
)
