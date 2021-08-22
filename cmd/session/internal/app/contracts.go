package app

import (
	"context"

	"github.com/gofrs/uuid"
)

type (
	// Repo interface for session data repository.
	Repo interface {
		// Save saves the new user session in a database.
		// Errors: unknown.
		Save(context.Context, Session) error
		// ByID returns user session by session id.
		// Errors: ErrNotFound, unknown.
		ByID(context.Context, uuid.UUID) (*Session, error)
		// Delete removes user session.
		// Errors: unknown.
		Delete(context.Context, uuid.UUID) error
	}

	// Auth interface for generate access and refresh token by subject.
	Auth interface {
		// Token generate tokens by subject with expire time.
		// Errors: unknown.
		Token(Subject) (*Token, error)
		// Subject unwrap Subject info from token.
		// Errors: ErrInvalidToken, ErrExpiredToken, unknown.
		Subject(token string) (*Subject, error)
	}

	// ID generator for session.
	ID interface {
		// New generate new ID for session.
		New() uuid.UUID
	}
)
