package repo

import "errors"

// Errors.
var (
	ErrNegativePosition = errors.New("wrong seek position <0")
	ErrUnexpectedWhence = errors.New("unexpected whence")
)
