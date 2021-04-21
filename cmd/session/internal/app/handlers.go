package app

import (
	"context"
	"fmt"
	"time"
)

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
