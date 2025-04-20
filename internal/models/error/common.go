package custom_errors

import "errors"

var (
	ErrBadRequest         = errors.New("bad request")
	ErrNotFound           = errors.New("not found")
	ErrMaxRetriesExceeded = errors.New("max retries count exceeded")
	ErrRequestTimeout     = errors.New("request timed out")
	ErrInternal           = errors.New("internal error")
)
