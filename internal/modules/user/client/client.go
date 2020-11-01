// Package client provide to internal method of service user.
package client

import (
	"context"
	"fmt"

	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/rpc/pb"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/app"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Client to user microservice.
type Client struct {
	conn pb.UserClient
}

// New build and returns new client to microservice user.
func New(conn grpc.ClientConnInterface) *Client {
	return &Client{conn: pb.NewUserClient(conn)}
}

// User contains main user info.
type User struct {
	ID    int
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
	res, err := c.conn.Access(ctx, &pb.RequestAccess{
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

	return &User{
		ID:    int(res.Id),
		Email: res.Email,
		Name:  res.Name,
	}, nil
}
