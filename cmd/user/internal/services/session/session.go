// Package session needed for get user session by token.
package session

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/gofrs/uuid"

	session "github.com/Meat-Hook/back-template/cmd/session/client"
	"github.com/Meat-Hook/back-template/cmd/user/internal/app"
)

var _ app.AuthSvc = &Client{}

// For easy testing.
type sessionSvc interface {
	Session(ctx context.Context, token string) (*session.Session, error)
	RemoveSession(ctx context.Context, sessionID uuid.UUID) error
	NewSession(ctx context.Context, userID uuid.UUID, ip net.IP, userAgent string) (*session.Token, error)
}

// Client wrapper for session microservice.
type Client struct {
	session sessionSvc
}

// New build and returns new session Client.
func New(svc sessionSvc) *Client {
	return &Client{session: svc}
}

// Session for implements app.AuthSvc.
func (c *Client) Session(ctx context.Context, token string) (*app.Session, error) {
	res, err := c.session.Session(ctx, token)
	switch {
	case errors.Is(err, session.ErrNotFound):
		return nil, app.ErrNotFound
	case err != nil:
		return nil, fmt.Errorf("c.session.Session: %w", err)
	}

	return &app.Session{
		ID:     res.ID,
		UserID: res.UserID,
	}, nil
}

// NewSession for implements app.AuthSvc.
func (c *Client) NewSession(ctx context.Context, userID uuid.UUID, origin app.Origin) (*app.Token, error) {
	res, err := c.session.NewSession(ctx, userID, origin.IP, origin.UserAgent)
	if err != nil {
		return nil, fmt.Errorf("c.session.NewSession: %w", err)
	}

	return &app.Token{Value: res.Value}, nil
}

// RemoveSession for implements app.AuthSvc.
func (c *Client) RemoveSession(ctx context.Context, sessionID uuid.UUID) error {
	err := c.session.RemoveSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("c.session.RemoveSession: %w", err)
	}

	return nil
}
