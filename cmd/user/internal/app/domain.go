package app

import (
	"encoding/json"
	"io"
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
		Avatars   []uuid.UUID
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

	// File contains file info.
	File struct {
		io.ReadCloser
		// ID file id.
		ID uuid.UUID
		// Size contains file size.
		Size int64
		// Metadata contains file meta info.
		Metadata json.RawMessage
	}
)
