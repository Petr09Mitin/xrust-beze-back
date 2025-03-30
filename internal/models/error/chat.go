package custom_errors

import "errors"

var (
	ErrInvalidMessage = errors.New("invalid message")
)
