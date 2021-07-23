// Package session needed for get user session by token.
package session

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofrs/uuid"

	session "github.com/Meat-Hook/back-template/cmd/session/client"
	"github.com/Meat-Hook/back-template/cmd/user/internal/app"
)

var _ app.Auth = &Client{}

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
		return nil, app.ErrNotFound
	case err != nil:
		return nil, fmt.Errorf("session: %w", err)
	}

	return &app.Session{
		ID:     res.ID,
		UserID: res.UserID,
	}, nil
}

func (c *Client) NewSession(ctx context.Context, userID uuid.UUID, origin app.Origin) (*app.Token, error) {
	panic("implement me")
}

func (c *Client) RemoveSession(ctx context.Context, sessionID uuid.UUID) error {
	panic("implement me")
}
