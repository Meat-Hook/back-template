// Package users needed for get user info by his email and pass.
package users

import (
	"context"
	"errors"
	"fmt"

	app2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/app"
	"github.com/Meat-Hook/back-template/internal/cmd/user/client"
)

var _ app2.Users = &Client{}

//go:generate mockgen -source=users.go -destination mock.app.contracts_test.go -package users_test

// For easy testing.
type userSvc interface {
	Access(ctx context.Context, email, pass string) (*client.User, error)
}

// Client wrapper for users microservice.
type Client struct {
	users userSvc
}

// New build and returns new user Client.
func New(svc userSvc) *Client {
	return &Client{users: svc}
}

// Access for implements app.Users.
func (c *Client) Access(ctx context.Context, email, password string) (*app2.User, error) {
	res, err := c.users.Access(ctx, email, password)
	switch {
	case errors.Is(err, client.ErrNotFound):
		return nil, app2.ErrNotFound
	case errors.Is(err, client.ErrNotValidPass):
		return nil, app2.ErrNotValidPassword
	case err != nil:
		return nil, fmt.Errorf("user access: %w", err)
	}

	return &app2.User{
		ID:    res.ID,
		Email: res.Email,
		Name:  res.Name,
	}, nil
}
