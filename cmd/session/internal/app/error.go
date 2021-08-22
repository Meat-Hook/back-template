package app

import (
	"errors"
)

// Errors.
var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidToken = errors.New("not valid auth")
)
