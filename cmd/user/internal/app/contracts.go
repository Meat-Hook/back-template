package app

import (
	"context"

	"github.com/gofrs/uuid"
)

type (
	// Repo interface for user data repository.
	Repo interface {
		// Save adds to the new user in repository.
		// Errors: ErrEmailExist, ErrUsernameExist, unknown.
		Save(context.Context, User) (uuid.UUID, error)
		// Update update user info.
		// Errors: ErrUsernameExist, ErrEmailExist, unknown.
		Update(context.Context, User) error
		// Delete removes user from repository by id.
		// Errors: unknown.
		Delete(context.Context, uuid.UUID) error
		// ByID returning user info by id.
		// Errors: ErrNotFound, unknown.
		ByID(context.Context, uuid.UUID) (*User, error)
		// ByEmail returning user info by email.
		// Errors: ErrNotFound, unknown.
		ByEmail(context.Context, string) (*User, error)
		// ByUsername returning user info by username.
		// Errors: ErrNotFound, unknown.
		ByUsername(context.Context, string) (*User, error)
		// ListUserByUsername returning list user info.
		// Errors: unknown.
		ListUserByUsername(context.Context, string, SearchParams) ([]User, int, error)
	}

	// Hasher module responsible for hashing password.
	Hasher interface {
		// Hashing returns the hashed version of the password.
		// Errors: unknown.
		Hashing(password string) ([]byte, error)
		// Compare compares two passwords for matches.
		Compare(hashedPassword []byte, password []byte) bool
	}

	// Auth module for get user session by token.
	Auth interface {
		// Session returns user session by his token.
		// Errors: ErrNotFound, unknown.
		Session(ctx context.Context, token string) (*Session, error)
		// NewSession generate new session for specific user.
		// Errors: unknown.
		NewSession(ctx context.Context, userID uuid.UUID, origin Origin) (*Token, error)
		// RemoveSession removes session by id.
		// Errors: ErrNotFound, unknown.
		RemoveSession(ctx context.Context, sessionID uuid.UUID) error
	}
)
