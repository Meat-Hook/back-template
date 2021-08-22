package app

import (
	"errors"
)

// Errors.
var (
	ErrNotFound   = errors.New("not found")
	ErrNotValidID = errors.New("not valid file id")
)
