package app

import (
	"context"
	"errors"
	"net"
	"strings"
)

// Errors.
var (
	ErrUnknownToken              = errors.New("unknown token")
	ErrEmailExist                = errors.New("email exist")
	ErrUsernameExist             = errors.New("username exist")
	ErrNotFound                  = errors.New("not found")
	ErrNotDifferent              = errors.New("the values must be different")
	ErrNotValidPassword          = errors.New("not valid password")
	ErrInvalidToken              = errors.New("not valid auth")
	ErrExpiredToken              = errors.New("auth is expired")
	ErrUsernameNeedDifferentiate = errors.New("username need to differentiate")
	ErrEmailNeedDifferentiate    = errors.New("email need to differentiate")
	ErrNotUnknownKindTask        = errors.New("unknown task kind")
	ErrCodeExpired               = errors.New("code is expired")
	ErrNotValidCode              = errors.New("code not equal")
)

type (
	// Repo interface for session data repository.
	Repo interface {
		// Save saves the new user session in a database.
		// Errors: unknown.
		Save(context.Context, Session) error
		// Session returns user session by session id.
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

	// Token contains access and refresh token.
	Token struct {
		Access  string
		Refresh string
	}

	// Subject contains info to be saved in token.
	Subject struct {
		SessionID string
	}

	// User contains user information.
	User struct {
		ID    int
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
		ID     string
		Origin Origin
		Token  Token
		UserID int
	}

	// Module contains business logic for user methods.
	Module struct {
		session Repo
		user    Users
		auth    Auth
		id      ID
	}
)

// Login generate new session and return user info.
func (m *Module) Login(ctx context.Context, email, password string, origin Origin) (*User, *Token, error) {
	email = strings.ToLower(email)

	user, err := m.user.Access(ctx, email, password)
	if err != nil {
		return nil, nil, err
	}

	sessionID := m.id.New()

	token, err := m.auth.Token(Subject{SessionID: sessionID})
	if err != nil {
		return nil, nil, err
	}

	session := Session{
		ID:     sessionID,
		Origin: origin,
		Token:  *token,
	}

	err = m.session.Save(ctx, session)
	if err != nil {
		return nil, nil, err
	}

	return user, token, nil
}

// Logout remove user session.
func (m *Module) Logout(ctx context.Context, auth Session) error {
	return m.session.Delete(ctx, auth.ID)
}

// Session get user session by access token.
func (m *Module) Session(ctx context.Context, token Token) (*Session, error) {
	subject, err := m.auth.Subject(token.Access)
	if err != nil {
		return nil, err
	}

	session, err := m.session.ByID(ctx, subject.SessionID)
	if err != nil {
		return nil, err
	}

	if token.Access != session.Token.Access {
		return nil, ErrUnknownToken
	}

	return session, nil
}

// Refresh access token by refresh token.
func (m *Module) Refresh(ctx context.Context, token Token, origin Origin) (*Token, error) {
	subject, err := m.auth.Subject(token.Refresh)
	if err != nil {
		return nil, err
	}

	session, err := m.session.ByID(ctx, subject.SessionID)
	if err != nil {
		return nil, err
	}

	err = m.session.Delete(ctx, session.ID)
	if err != nil {
		return nil, err
	}

	sessionID := m.id.New()
	newToken, err := m.auth.Token(Subject{SessionID: sessionID})
	if err != nil {
		return nil, err
	}

	newSession := Session{
		ID:     sessionID,
		Origin: origin,
		Token:  *newToken,
	}

	err = m.session.Save(ctx, newSession)
	if err != nil {
		return nil, err
	}

	return newToken, nil
}
