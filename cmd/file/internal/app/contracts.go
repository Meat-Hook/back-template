package app

import (
	"context"
	"encoding/json"
	"io"

	"github.com/gofrs/uuid"
)

type (
	// Repo interface for session data repository.
	Repo interface {
		// Save saves the new file to database.
		// Errors: unknown.
		Save(context.Context, io.Reader) (uuid.UUID, error)
		// Read returns file by id.
		// Errors: ErrNotFound, unknown.
		Read(context.Context, uuid.UUID) (*File, error)
		// SetMetadata set the file metadata.
		// Errors: ErrNotFound, unknown.
		SetMetadata(context.Context, uuid.UUID, json.RawMessage) error
		// Delete removes file with metadata from database.
		// Errors: unknown.
		Delete(context.Context, uuid.UUID) error
	}
)
