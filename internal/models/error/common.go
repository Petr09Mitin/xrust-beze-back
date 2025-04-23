package custom_errors

import (
	"errors"
	"fmt"
)

var (
	ErrBadRequest         = errors.New("bad request")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrNotFound           = errors.New("not found")
	ErrMaxRetriesExceeded = errors.New("max retries count exceeded")
	ErrRequestTimeout     = errors.New("request timed out")
	ErrInternal           = errors.New("internal error")
	ErrInvalidIDType      = fmt.Errorf("%w: the provided hex string is not a valid ObjectID", ErrBadRequest)
)
