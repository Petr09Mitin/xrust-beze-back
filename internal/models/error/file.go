package custom_errors

import "errors"

var (
	ErrFileNotFound      = errors.New("file not found")
	ErrInvalidFileFormat = errors.New("invalid file format")
)
