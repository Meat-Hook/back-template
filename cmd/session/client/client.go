// Package client provide to internal method of service session.
package client

import (
	"context"
	"fmt"

	app2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/app"
	log2 "github.com/Meat-Hook/back-template/internal/libs/log"
	pb "github.com/Meat-Hook/back-template/proto/gen/go/session/v1"
	"github.com/gofrs/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Client to session microservice.
type Client struct {
	conn pb.SessionServiceClient
}

// New build and returns new client to microservice session.
func New(conn grpc.ClientConnInterface) *Client {
	return &Client{conn: pb.NewSessionServiceClient(conn)}
}

// Session contains main session info.
type Session struct {
	ID     string
	UserID uuid.UUID
}

// Errors.
var (
	ErrNotFound = app2.ErrNotFound
)

// Session get user session by his auth token.
func (c *Client) Session(ctx context.Context, token string) (*Session, error) {
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		log2.ReqID: []string{log2.ReqIDFromCtx(ctx)},
	})

	res, err := c.conn.Session(ctx, &pb.SessionRequest{
		Token: token,
	})
	switch {
	case status.Code(err) == codes.NotFound:
		return nil, fmt.Errorf("%w: %s", ErrNotFound, err)
	case err != nil:
		return nil, fmt.Errorf("session: %w", err)
	}

	uid, err := uuid.FromString(res.UserId)
	if err != nil {
		return nil, fmt.Errorf("parse uuid: %w", err)
	}

	return &Session{
		ID:     res.Id,
		UserID: uid,
	}, nil
}
