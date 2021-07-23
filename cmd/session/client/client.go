// Package client provide to internal method of service session.
package client

import (
	"context"
	"fmt"
	"net"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/Meat-Hook/back-template/cmd/session/internal/app"
	"github.com/Meat-Hook/back-template/libs/log"
	pb "github.com/Meat-Hook/back-template/proto/gen/go/session/v1"
)

// Errors.
var (
	ErrNotFound = app.ErrNotFound
)

// Client to session microservice.
type Client struct {
	conn pb.ServiceClient
}

// New build and returns new client to microservice session.
func New(conn grpc.ClientConnInterface) *Client {
	return &Client{conn: pb.NewServiceClient(conn)}
}

// Session contains main session info.
type Session struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

// Token contains user's authorization token.
type Token struct {
	Value string
}

// Session get user session by his auth token.
func (c *Client) Session(ctx context.Context, token string) (*Session, error) {
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		log.ReqID: []string{log.ReqIDFromCtx(ctx)},
	})

	res, err := c.conn.Session(ctx, &pb.SessionRequest{
		Token: token,
	})
	switch {
	case status.Code(err) == codes.NotFound:
		return nil, fmt.Errorf("%w: %s", ErrNotFound, err)
	case err != nil:
		return nil, fmt.Errorf("c.conn.Session: %w", err)
	}

	userUID, err := uuid.FromString(res.UserId.Value)
	if err != nil {
		return nil, fmt.Errorf("uuid.FromString: %w", err)
	}

	sessionUID, err := uuid.FromString(res.SessionId.Value)
	if err != nil {
		return nil, fmt.Errorf("uuid.FromString: %w", err)
	}

	return &Session{
		ID:     sessionUID,
		UserID: userUID,
	}, nil
}

// RemoveSession remove user session by session ID.
func (c *Client) RemoveSession(ctx context.Context, sessionID uuid.UUID) error {
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		log.ReqID: []string{log.ReqIDFromCtx(ctx)},
	})

	_, err := c.conn.RemoveSession(ctx, &pb.RemoveSessionRequest{
		SessionId: &pb.UUID{Value: sessionID.String()},
	})
	if err != nil {
		return fmt.Errorf("c.conn.RemoveSession: %w", err)
	}

	return nil
}

// NewSession make new session for user.
func (c *Client) NewSession(ctx context.Context, userID uuid.UUID, ip net.IP, userAgent string) (*Token, error) {
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		log.ReqID: []string{log.ReqIDFromCtx(ctx)},
	})

	res, err := c.conn.NewSession(ctx, &pb.NewSessionRequest{
		UserId:    &pb.UUID{Value: userID.String()},
		Ip:        ip.String(),
		UserAgent: userAgent,
	})
	if err != nil {
		return nil, fmt.Errorf("c.conn.NewSession: %w", err)
	}

	return &Token{Value: res.Token}, nil
}
