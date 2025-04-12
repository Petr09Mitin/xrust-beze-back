package custom_errors

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidMessage             = errors.New("invalid message")
	ErrInvalidMessageEvent        = fmt.Errorf("%w: invalid event", ErrInvalidMessage)
	ErrBroadcastingTextMessage    = errors.New("error broadcasting text message")
	ErrInvalidMessageType         = errors.New("invalid message type")
	ErrNoChannelID                = errors.New("no channel id")
	ErrNoMessageID                = errors.New("no message id")
	ErrStructurizationUnavailable = errors.New("structurization is temporary unavailable, try again later")
)
