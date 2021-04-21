package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/gofrs/uuid"
)

// Errors.
var (
	ErrNotFound         = errors.New("not found")
	ErrNotValidPassword = errors.New("not valid password")
	ErrInvalidToken     = errors.New("not valid auth")
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

	// Module contains business logic for user methods.
	Module struct {
		session Repo
		user    Users
		auth    Auth
		id      ID
	}
)

// New build and returns new session module.
func New(r Repo, u Users, a Auth, id ID) *Module {
	return &Module{
		session: r,
		user:    u,
		auth:    a,
		id:      id,
	}
}

// Login generate new session and return user info.
func (m *Module) Login(ctx context.Context, email, password string, origin Origin) (*User, *Token, error) {
	user, err := m.user.Access(ctx, email, password)
	if err != nil {
		return nil, nil, fmt.Errorf("user access: %w", err)
	}

	sessionID := m.id.New()

	token, err := m.auth.Token(Subject{SessionID: sessionID})
	if err != nil {
		return nil, nil, fmt.Errorf("auth token: %w", err)
	}

	session := Session{
		ID:        sessionID,
		Origin:    origin,
		Token:     *token,
		UserID:    user.ID,
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}

	err = m.session.Save(ctx, session)
	if err != nil {
		return nil, nil, fmt.Errorf("session save: %w", err)
	}

	return user, token, nil
}

// Logout remove user session.
func (m *Module) Logout(ctx context.Context, session Session) error {
	return m.session.Delete(ctx, session.ID)
}

// Session get user session by access token.
func (m *Module) Session(ctx context.Context, token string) (*Session, error) {
	subject, err := m.auth.Subject(token)
	if err != nil {
		return nil, fmt.Errorf("auth subject: %w", err)
	}

	session, err := m.session.ByID(ctx, subject.SessionID)
	if err != nil {
		return nil, fmt.Errorf("session by id: %w", err)
	}

	return session, nil
}
