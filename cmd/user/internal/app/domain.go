package app

import (
	"time"

	"github.com/gofrs/uuid"
)

type (
	// SearchParams params for search users.
	SearchParams struct {
		Limit  uint
		Offset uint
	}

	// Session contains user session information.
	Session struct {
		ID     string
		UserID uuid.UUID
	}

	// User contains user information.
	User struct {
		ID        uuid.UUID
		Email     string
		Name      string
		PassHash  []byte
		CreatedAt time.Time
		UpdatedAt time.Time
	}
)
