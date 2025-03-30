package custom_errors

import "errors"

var (
	ErrNotFound = errors.New("not found")
)
