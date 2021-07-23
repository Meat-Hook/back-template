package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
)

// VerificationEmail check exists or not user email.
func (m *Module) VerificationEmail(ctx context.Context, email string) error {
	_, err := m.user.ByEmail(ctx, email)
	switch {
	case errors.Is(err, ErrNotFound):
		return nil
	case err == nil:
		return ErrEmailExist
	default:
		return fmt.Errorf("m.user.ByEmail: %w", err)
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
		return fmt.Errorf("m.user.ByUsername: %w", err)
	}
}

// CreateUser create new user by params.
func (m *Module) CreateUser(ctx context.Context, email, username, password string) (uuid.UUID, error) {
	passHash, err := m.hash.Hashing(password)
	if err != nil {
		return uuid.Nil, fmt.Errorf("m.hash.Hashing: %w", err)
	}
	email = strings.ToLower(email)

	newUser := User{
		Email:    email,
		Name:     username,
		PassHash: passHash,
	}

	userID, err := m.user.Save(ctx, newUser)
	if err != nil {
		return uuid.Nil, fmt.Errorf("m.user.Save: %w", err)
	}

	return userID, nil
}

// UserByID get user by id.
func (m *Module) UserByID(ctx context.Context, _ Session, userID uuid.UUID) (*User, error) {
	return m.user.ByID(ctx, userID)
}

// DeleteUser remove user from db.
func (m *Module) DeleteUser(ctx context.Context, session Session) error {
	return m.user.Delete(ctx, session.UserID)
}

// UpdateUsername update username.
func (m *Module) UpdateUsername(ctx context.Context, session Session, username string) error {
	user, err := m.user.ByID(ctx, session.UserID)
	if err != nil {
		return fmt.Errorf("m.user.ByID: %w", err)
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
		return fmt.Errorf("m.user.ByID: %w", err)
	}

	if !m.hash.Compare(user.PassHash, []byte(oldPass)) {
		return ErrNotValidPassword
	}

	if m.hash.Compare(user.PassHash, []byte(newPass)) {
		return ErrNotDifferent
	}

	passHash, err := m.hash.Hashing(newPass)
	if err != nil {
		return fmt.Errorf("m.hash.Hashing: %w", err)
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

// Login make new session and returns auth token.
func (m *Module) Login(ctx context.Context, email, password string, origin Origin) (*Token, error) {
	email = strings.ToLower(email)
	user, err := m.user.ByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("m.user.ByEmail: %w", err)
	}

	if !m.hash.Compare(user.PassHash, []byte(password)) {
		return nil, ErrNotValidPassword
	}

	return m.auth.NewSession(ctx, user.ID, origin)
}

// Logout remove user session.
func (m *Module) Logout(ctx context.Context, session Session) error {
	return m.auth.RemoveSession(ctx, session.ID)
}
