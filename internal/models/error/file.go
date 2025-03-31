package custom_errors

import "errors"

var (
	ErrFileNotFound = errors.New("file not found")
)
