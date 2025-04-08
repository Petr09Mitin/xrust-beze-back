package custom_errors

import "errors"

var (
	ErrWrongPassword = errors.New("wrong password (can't authorize)")
)
