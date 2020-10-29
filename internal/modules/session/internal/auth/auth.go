// Package auth contains methods for working with authorization tokens,
// their generation and parsing.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/Meat-Hook/back-template/internal/modules/session/internal/app"
	"github.com/o1egl/paseto/v2"
	"github.com/rs/xid"
)

var _ app.Auth = &Auth{}

const (
	accessExp  = time.Minute * 30
	refreshExp = time.Hour * 24 * 365
)

const (
	iss = `session-service`

	access  = `-access`
	refresh = `-refresh`
)

// Errors.
var (
	ErrValidateAlg = errors.New("unexpected signing method")
)

// Auth is an implements app.Auth.
// Responsible for working with authorization tokens, be it cookies or jwt.
type Auth struct {
	key []byte
}

// New creates and returns new instance auth.
func New(jwtKey string) *Auth {
	return &Auth{
		key: []byte(jwtKey),
	}
}

// Token need for implements app.Auth.
func (a *Auth) Token(subject app.Subject) (*app.Token, error) {
	jsonAccessToken := paseto.JSONToken{
		Audience:   "",
		Issuer:     iss,
		Jti:        xid.New().String() + access,
		Subject:    subject.SessionID,
		Expiration: time.Now().Add(accessExp),
		IssuedAt:   time.Now(),
		NotBefore:  time.Now(),
	}

	accessToken, err := paseto.Encrypt(a.key, jsonAccessToken, "")
	if err != nil {
		return nil, fmt.Errorf("encrypt access token: %w", err)
	}

	jsonRefreshToken := paseto.JSONToken{
		Audience:   "",
		Issuer:     iss,
		Jti:        xid.New().String() + refresh,
		Subject:    subject.SessionID,
		Expiration: time.Now().Add(accessExp),
		IssuedAt:   time.Now(),
		NotBefore:  time.Now(),
	}

	refreshToken, err := paseto.Encrypt(a.key, jsonRefreshToken, "")
	if err != nil {
		return nil, fmt.Errorf("encrypt refresh token: %w", err)
	}

	res := &app.Token{
		Access:  accessToken,
		Refresh: refreshToken,
	}

	return res, nil
}

// Subject need for implements app.Auth.
func (a *Auth) Subject(token string) (*app.Subject, error) {
	jsonToken := paseto.JSONToken{}

	err := paseto.Decrypt(token, a.key, &jsonToken, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", app.ErrInvalidToken, err)
	}

	if jsonToken.Expiration.Before(time.Now()) {
		return nil, app.ErrExpiredToken
	}

	sub := &app.Subject{
		SessionID: jsonToken.Subject,
	}

	return sub, nil
}
