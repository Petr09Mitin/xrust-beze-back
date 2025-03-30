package custom_errors

import "errors"

var (
	ErrInvalidMessage          = errors.New("invalid message")
	ErrBroadcastingTextMessage = errors.New("error broadcasting text message")
	ErrInvalidMessageType      = errors.New("invalid message type")
	ErrNoChannelID             = errors.New("no channel id")
)
