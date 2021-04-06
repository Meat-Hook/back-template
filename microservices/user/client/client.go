// Package client provide to internal method of service user.
package client

import (
	"context"
	"fmt"

	"github.com/Meat-Hook/back-template/libs/log"
	"github.com/Meat-Hook/back-template/microservices/user/internal/app"
	pb "github.com/Meat-Hook/back-template/proto/go/user/v1"
	"github.com/gofrs/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Client to user microservice.
type Client struct {
	conn pb.UserServiceClient
}

// New build and returns new client to microservice user.
func New(conn grpc.ClientConnInterface) *Client {
	return &Client{conn: pb.NewUserServiceClient(conn)}
}

// User contains main user info.
type User struct {
	ID    uuid.UUID
	Email string
	Name  string
}

// Errors.
var (
	ErrNotFound     = app.ErrNotFound
	ErrNotValidPass = app.ErrNotValidPassword
)

// Access get user info by his email and pass.
// Needed for user auth.
func (c *Client) Access(ctx context.Context, email, pass string) (*User, error) {
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		log.ReqID: []string{log.ReqIDFromCtx(ctx)},
	})

	res, err := c.conn.Access(ctx, &pb.AccessRequest{
		Email:    email,
		Password: pass,
	})
	switch {
	case status.Code(err) == codes.NotFound:
		return nil, fmt.Errorf("%w: %s", ErrNotFound, err)
	case status.Code(err) == codes.InvalidArgument:
		return nil, fmt.Errorf("%w: %s", ErrNotValidPass, err)
	case err != nil:
		return nil, fmt.Errorf("access: %w", err)
	}

	uid, err := uuid.FromString(res.Id)
	if err != nil {
		return nil, fmt.Errorf("parse uuid: %w", err)
	}

	return &User{
		ID:    uid,
		Email: res.Email,
		Name:  res.Name,
	}, nil
}
