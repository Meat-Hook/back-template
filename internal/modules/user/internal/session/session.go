// Package session needed for get user session by token.
package session

import (
	"context"
	"errors"
	"fmt"

	session "github.com/Meat-Hook/back-template/internal/modules/session/client"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/app"
)

var _ app.Auth = &Client{}

//go:generate mockgen -source=session.go -destination mock.app.contracts_test.go -package session_test

// For easy testing.
type sessionSvc interface {
	Session(ctx context.Context, token string) (*session.Session, error)
}

// Client wrapper for session microservice.
type Client struct {
	session sessionSvc
}

// New build and returns new session Client.
func New(svc sessionSvc) *Client {
	return &Client{session: svc}
}

// Session for implements app.Auth.
func (c *Client) Session(ctx context.Context, token string) (*app.Session, error) {
	res, err := c.session.Session(ctx, token)
	switch {
	case errors.Is(err, session.ErrNotFound):
		return nil, fmt.Errorf("%w: %s", app.ErrNotFound, err)
	case err != nil:
		return nil, fmt.Errorf("session: %w", err)
	}

	return &app.Session{
		ID:     res.ID,
		UserID: res.UserID,
	}, nil
}
