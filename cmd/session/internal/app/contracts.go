package app

import (
	"context"
)

type (
	// Repo interface for session data repository.
	Repo interface {
		// Save saves the new user session in a database.
		// Errors: unknown.
		Save(context.Context, Session) error
		// ByID returns user session by session id.
		// Errors: ErrNotFound, unknown.
		ByID(context.Context, string) (*Session, error)
		// Delete removes user session.
		// Errors: unknown.
		Delete(ctx context.Context, sessionID string) error
	}

	// Users microservice for get user information.
	Users interface {
		// Access get user by email and check password.
		// Errors: ErrNotFound, ErrNotValidPassword, unknown.
		Access(ctx context.Context, email, password string) (*User, error)
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
		New() string
	}
)
