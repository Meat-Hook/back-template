package app

import (
	"net"
	"time"

	"github.com/gofrs/uuid"
)

type (
	// Token contains auth token.
	Token struct {
		Value string
	}

	// Subject contains info to be saved in token.
	Subject struct {
		SessionID string
	}

	// User contains user information.
	User struct {
		ID    uuid.UUID
		Email string
		Name  string
	}

	// Origin information about req user.
	Origin struct {
		IP        net.IP
		UserAgent string
	}

	// Session contains session info for identify a user.
	Session struct {
		ID        string
		Origin    Origin
		Token     Token
		UserID    uuid.UUID
		CreatedAt time.Time
		UpdatedAt time.Time
	}
)
