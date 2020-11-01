// Package auth contains methods for working with authorization tokens,
// their generation and parsing.
package auth

import (
	"fmt"

	"github.com/Meat-Hook/back-template/internal/modules/session/internal/app"
	"github.com/o1egl/paseto/v2"
)

var _ app.Auth = &Auth{}

// Auth is an implements app.Auth.
// Responsible for working with authorization tokens, be it cookies or jwt.
type Auth struct {
	key []byte
}

// New creates and returns new instance auth.
func New(secretKey string) *Auth {
	return &Auth{
		key: []byte(secretKey),
	}
}

type jsonToken struct {
	SessionID string `json:"session_id"`
}

// Token need for implements app.Auth.
func (a *Auth) Token(subject app.Subject) (*app.Token, error) {
	t := jsonToken{
		SessionID: subject.SessionID,
	}

	value, err := paseto.Encrypt(a.key, t, "")
	if err != nil {
		return nil, fmt.Errorf("encrypt access token: %w", err)
	}

	res := &app.Token{
		Value: value,
	}

	return res, nil
}

// Subject need for implements app.Auth.
func (a *Auth) Subject(token string) (*app.Subject, error) {
	t := jsonToken{
		SessionID: token,
	}

	err := paseto.Decrypt(token, a.key, &t, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", app.ErrInvalidToken, err)
	}

	sub := &app.Subject{
		SessionID: t.SessionID,
	}

	return sub, nil
}
