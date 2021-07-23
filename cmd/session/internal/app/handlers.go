package app

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
)

// RemoveSession remove user session.
func (m *Module) RemoveSession(ctx context.Context, sessionID uuid.UUID) error {
	return m.session.Delete(ctx, sessionID)
}

// NewSession save new user session.
func (m *Module) NewSession(ctx context.Context, userID uuid.UUID, origin Origin) (*Token, error) {
	sessionID := m.id.New()
	token, err := m.auth.Token(Subject{SessionID: sessionID})
	if err != nil {
		return nil, fmt.Errorf("m.auth.Token: %w", err)
	}

	session := Session{
		ID:        sessionID,
		Origin:    origin,
		Token:     *token,
		UserID:    userID,
		CreatedAt: time.Time{}, // Will set in database.
		UpdatedAt: time.Time{}, // Will set in database.
	}

	err = m.session.Save(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("m.session.Save: %w", err)
	}

	return token, nil
}

// Session get user session by access token.
func (m *Module) Session(ctx context.Context, token string) (*Session, error) {
	subject, err := m.auth.Subject(token)
	if err != nil {
		return nil, fmt.Errorf("m.auth.Subject: %w", err)
	}

	session, err := m.session.ByID(ctx, subject.SessionID)
	if err != nil {
		return nil, fmt.Errorf("m.session.ByID: %w", err)
	}

	return session, nil
}
