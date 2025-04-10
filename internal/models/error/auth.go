package custom_errors

import "errors"

var (
	ErrWrongPassword      = errors.New("wrong password (can't authorize)")
	ErrMissingUserID      = errors.New("user_id not found in context")
	ErrInvalidUserIDType  = errors.New("user_id has invalid type")
	ErrUserIDMismatch     = errors.New("cannot update another user's data")
	ErrInvalidUserID      = errors.New("invalid user_id format")
	ErrMissingLoginField  = errors.New("enter email or username")
	ErrTooManyLoginFields = errors.New("enter only one: email or username")
)
