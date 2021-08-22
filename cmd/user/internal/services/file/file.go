// Package file needed for manage user avatars.
package file

import (
	"context"
	"fmt"
	"io"

	"github.com/gofrs/uuid"

	"github.com/Meat-Hook/back-template/cmd/user/internal/app"
)

var _ app.FileSvc = &Client{}

// For easy testing.
type fileSvc interface {
	Upload(ctx context.Context, r io.Reader) (uuid.UUID, error)
	Delete(ctx context.Context, fileID uuid.UUID) error
}

// Client wrapper for session microservice.
type Client struct {
	file fileSvc
}

// New build and returns new session Client.
func New(svc fileSvc) *Client {
	return &Client{file: svc}
}

// Upload for implements app.AuthSvc.
func (c *Client) Upload(ctx context.Context, file io.Reader) (uuid.UUID, error) {
	res, err := c.file.Upload(ctx, file)
	if err != nil {
		return uuid.Nil, fmt.Errorf("c.file.Upload: %w", err)
	}

	return res, nil
}

// Delete for implements app.AuthSvc.
func (c *Client) Delete(ctx context.Context, fileID uuid.UUID) error {
	err := c.file.Delete(ctx, fileID)
	if err != nil {
		return fmt.Errorf("c.file.Delete: %w", err)
	}

	return nil
}
