package custom_errors

import "errors"

var (
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrUserNotExists         = errors.New("user does not exist")
	ErrMissingPassword       = errors.New("password field missing or not a string")
	ErrValidationFailed      = errors.New("validation failed: one or more fields are invalid")
)
