// Package client provide to internal method of service user.
package client

import (
	"context"
	"fmt"

	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/rpc/pb"
	"google.golang.org/grpc"
)

// Client to user microservice.
type Client struct {
	conn pb.UserClient
}

// New build and returns new client to microservice user.
func New(conn *grpc.ClientConn) (*Client, error) {
	return &Client{conn: pb.NewUserClient(conn)}, nil
}

// User contains main user info.
type User struct {
	ID    int
	Email string
	Name  string
}

// Access get user info by his email and pass.
// Needed for user auth.
func (c *Client) Access(ctx context.Context, email, pass string) (*User, error) {
	res, err := c.conn.Access(ctx, &pb.RequestAccess{
		Email:    email,
		Password: pass,
	})
	if err != nil {
		return nil, fmt.Errorf("access: %w", err)
	}

	return &User{
		ID:    int(res.Id),
		Email: res.Email,
		Name:  res.Name,
	}, nil
}
