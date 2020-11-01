// Package users needed for get user info by his email and pass.
package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/Meat-Hook/back-template/internal/modules/session/internal/app"
	user "github.com/Meat-Hook/back-template/internal/modules/user/client"
)

var _ app.Users = &Client{}

//go:generate mockgen -source=users.go -destination mock.app.contracts_test.go -package users_test

// For easy testing.
type userSvc interface {
	Access(ctx context.Context, email, pass string) (*user.User, error)
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
func (c *Client) Access(ctx context.Context, email, password string) (*app.User, error) {
	res, err := c.users.Access(ctx, email, password)
	switch {
	case errors.Is(err, user.ErrNotFound):
		return nil, fmt.Errorf("%w: %s", app.ErrNotFound, err)
	case errors.Is(err, user.ErrNotValidPass):
		return nil, fmt.Errorf("%w: %s", app.ErrNotValidPassword, err)
	case err != nil:
		return nil, fmt.Errorf("session: %w", err)
	}

	return &app.User{
		ID:    res.ID,
		Email: res.Email,
		Name:  res.Name,
	}, nil
}
