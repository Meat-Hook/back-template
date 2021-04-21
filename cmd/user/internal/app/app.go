// Package app contains all logic of the microservice.
package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

// Errors.
var (
	ErrEmailExist       = errors.New("email exist")
	ErrUsernameExist    = errors.New("username exist")
	ErrNotFound         = errors.New("not found")
	ErrNotDifferent     = errors.New("the values must be different")
	ErrNotValidPassword = errors.New("not valid password")
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
	}

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

	// Module contains business logic for user methods.
	Module struct {
		user         Repo
		hash         Hasher
		auth         Auth
	}
)

// New build and returns new Module for working with user info.
func New(r Repo, h Hasher, a Auth) *Module {
	return &Module{
		user:         r,
		hash:         h,
		auth:         a,
	}
}

// VerificationEmail check exists or not user email.
func (m *Module) VerificationEmail(ctx context.Context, email string) error {
	_, err := m.user.ByEmail(ctx, email)
	switch {
	case errors.Is(err, ErrNotFound):
		return nil
	case err == nil:
		return ErrEmailExist
	default:
		return fmt.Errorf("user by email: %w", err)
	}
}

// VerificationUsername check exists or not username.
func (m *Module) VerificationUsername(ctx context.Context, username string) error {
	_, err := m.user.ByUsername(ctx, username)
	switch {
	case errors.Is(err, ErrNotFound):
		return nil
	case err == nil:
		return ErrUsernameExist
	default:
		return fmt.Errorf("user by username: %w", err)
	}
}

// CreateUser create new user by params.
func (m *Module) CreateUser(ctx context.Context, email, username, password string) (uuid.UUID, error) {
	passHash, err := m.hash.Hashing(password)
	if err != nil {
		return uuid.Nil, fmt.Errorf("hash hashing: %w", err)
	}
	email = strings.ToLower(email)

	newUser := User{
		Email:    email,
		Name:     username,
		PassHash: passHash,
	}

	userID, err := m.user.Save(ctx, newUser)
	if err != nil {
		return uuid.Nil, fmt.Errorf("user save: %w", err)
	}

	return userID, nil
}

// UserByID get user by id.
func (m *Module) UserByID(ctx context.Context, _ Session, userID uuid.UUID) (*User, error) {
	return m.user.ByID(ctx, userID)
}

// DeleteUser remove user from repo.
func (m *Module) DeleteUser(ctx context.Context, session Session) error {
	return m.user.Delete(ctx, session.UserID)
}

// UpdateUsername update username.
func (m *Module) UpdateUsername(ctx context.Context, session Session, username string) error {
	user, err := m.user.ByID(ctx, session.UserID)
	if err != nil {
		return fmt.Errorf("user by id: %w", err)
	}

	if user.Name == username {
		return ErrNotDifferent
	}
	user.Name = username

	return m.user.Update(ctx, *user)
}

// UpdatePassword update user password.
func (m *Module) UpdatePassword(ctx context.Context, session Session, oldPass, newPass string) error {
	user, err := m.user.ByID(ctx, session.UserID)
	if err != nil {
		return fmt.Errorf("user by id: %w", err)
	}

	if !m.hash.Compare(user.PassHash, []byte(oldPass)) {
		return ErrNotValidPassword
	}

	if m.hash.Compare(user.PassHash, []byte(newPass)) {
		return ErrNotDifferent
	}

	passHash, err := m.hash.Hashing(newPass)
	if err != nil {
		return fmt.Errorf("hash hashing: %w", err)
	}
	user.PassHash = passHash

	return m.user.Update(ctx, *user)
}

// ListUserByUsername get users by username.
func (m *Module) ListUserByUsername(ctx context.Context, _ Session, username string, p SearchParams) ([]User, int, error) {
	return m.user.ListUserByUsername(ctx, username, p)
}

// Auth get user session by token.
func (m *Module) Auth(ctx context.Context, token string) (*Session, error) {
	return m.auth.Session(ctx, token)
}

// Access finds a user by email and compares his password to allow access.
func (m *Module) Access(ctx context.Context, email, password string) (*User, error) {
	email = strings.ToLower(email)
	user, err := m.user.ByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user by email: %w", err)
	}

	if !m.hash.Compare(user.PassHash, []byte(password)) {
		return nil, ErrNotValidPassword
	}

	return user, nil
}
