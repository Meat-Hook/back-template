package app

import (
	"errors"
)

// Errors.
var (
	ErrEmailExist       = errors.New("email exist")
	ErrUsernameExist    = errors.New("username exist")
	ErrNotFound         = errors.New("not found")
	ErrNotDifferent     = errors.New("the values must be different")
	ErrNotValidPassword = errors.New("not valid password")
)
