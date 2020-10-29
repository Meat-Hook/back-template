package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/Meat-Hook/back-template/internal/modules/session/internal/api/rpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client struct {
	conn pb.SessionClient
}

type Config struct {
	Host string
	Port int
}

func New(conn *grpc.ClientConn) (*Client, error) {
	return &Client{conn: pb.NewSessionClient(conn)}, nil
}

type Session struct {
	ID     string
	UserID int
}

var (
	ErrNotFound = errors.New("not found")
)

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
