package custom_errors

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	ErrorsToHTTPStatusCodes = map[error]int{
		// common
		ErrNotFound:     http.StatusNotFound,
		ErrUnauthorized: http.StatusUnauthorized,
		ErrBadRequest:   http.StatusBadRequest,

		// auth
		ErrWrongPassword:      http.StatusBadRequest,
		ErrMissingUserID:      http.StatusUnauthorized,
		ErrInvalidUserIDType:  http.StatusBadRequest,
		ErrUserIDMismatch:     http.StatusForbidden,
		ErrInvalidUserID:      http.StatusBadRequest,
		ErrMissingLoginField:  http.StatusBadRequest,
		ErrTooManyLoginFields: http.StatusBadRequest,

		// chat
		ErrInvalidMessage:      http.StatusBadRequest,
		ErrInvalidMessageEvent: http.StatusBadRequest,
		ErrNoChannelID:         http.StatusBadRequest,
		ErrNoMessageID:         http.StatusBadRequest,

		// file
		ErrFileNotFound:      http.StatusNotFound,
		ErrInvalidFileFormat: http.StatusBadRequest,

		// user
		ErrEmailAlreadyExists:    http.StatusBadRequest,
		ErrUsernameAlreadyExists: http.StatusBadRequest,
		ErrUserNotExists:         http.StatusBadRequest,
		ErrMissingPassword:       http.StatusBadRequest,
	}
)

type HTTPError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"error"`
}

func (e HTTPError) Error() string {
	return e.Message
}

func getRootError(err error) error {
	currentErr := err
	for errors.Unwrap(currentErr) != nil {
		currentErr = errors.Unwrap(currentErr)
	}
	return currentErr
}

func MapErrorToHTTPError(err error) *HTTPError {
	if err == nil {
		return nil
	}

	baseErr := getRootError(err)

	code, ok := ErrorsToHTTPStatusCodes[baseErr]
	if !ok {
		code = http.StatusInternalServerError
	}

	return &HTTPError{
		StatusCode: code,
		Message:    err.Error(),
	}
}

func WriteHTTPError(c *gin.Context, err error) {
	var customErr *HTTPError
	if err == nil {
		customErr = MapErrorToHTTPError(ErrInternal)
	} else {
		customErr = MapErrorToHTTPError(err)
	}

	c.JSON(customErr.StatusCode, customErr)
}
