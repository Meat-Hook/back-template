package app

import (
	"errors"
)

// Errors.
var (
	ErrNotFound         = errors.New("not found")
	ErrNotValidPassword = errors.New("not valid password")
	ErrInvalidToken     = errors.New("not valid auth")
)
