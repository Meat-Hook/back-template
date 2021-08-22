package app

import (
	"context"
	"encoding/json"
	"io"

	"github.com/gofrs/uuid"
)

// UploadFile upload new file.
func (m *Module) UploadFile(ctx context.Context, file io.Reader) (uuid.UUID, error) {
	return m.file.Save(ctx, file)
}

// GetFile file from database.
func (m *Module) GetFile(ctx context.Context, fileID uuid.UUID) (*File, error) {
	return m.file.Read(ctx, fileID)
}

// SetMetadata set file metadata.
func (m *Module) SetMetadata(ctx context.Context, fileID uuid.UUID, metadata json.RawMessage) error {
	return m.file.SetMetadata(ctx, fileID, metadata)
}

// Delete file.
func (m *Module) Delete(ctx context.Context, fileID uuid.UUID) error {
	return m.file.Delete(ctx, fileID)
}
