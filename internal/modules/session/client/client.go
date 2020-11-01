// Package client provide to internal method of service session.
package client

import (
	"context"
	"fmt"

	"github.com/Meat-Hook/back-template/internal/modules/session/internal/api/rpc/pb"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/app"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Client to session microservice.
type Client struct {
	conn pb.SessionClient
}

// New build and returns new client to microservice session.
func New(conn grpc.ClientConnInterface) *Client {
	return &Client{conn: pb.NewSessionClient(conn)}
}

// Session contains main session info.
type Session struct {
	ID     string
	UserID int
}

// Errors.
var (
	ErrNotFound = app.ErrNotFound
)

// Session get user session by his auth token.
func (c *Client) Session(ctx context.Context, token string) (*Session, error) {
	res, err := c.conn.Session(ctx, &pb.RequestSession{
		Token: token,
	})
	switch {
	case status.Code(err) == codes.NotFound:
		return nil, fmt.Errorf("%w: %s", ErrNotFound, err)
	case err != nil:
		return nil, fmt.Errorf("session: %w", err)
	}

	return &Session{
		ID:     res.ID,
		UserID: int(res.UserID),
	}, nil
}
