package app

import (
	"net"
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
		ID     uuid.UUID
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

	// Token contains auth token.
	Token struct {
		Value string
	}

	// Subject contains info to be saved in token.
	Subject struct {
		SessionID uuid.UUID
	}

	// Origin information about req user.
	Origin struct {
		IP        net.IP
		UserAgent string
	}
)
