package custom_errors

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrMaxRetriesExceeded = errors.New("max retries count exceeded")
	ErrRequestTimeout     = errors.New("request timed out")
)
