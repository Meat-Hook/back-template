package app

import (
	"encoding/json"
	"io"

	"github.com/gofrs/uuid"
)

type (
	// File contains file info.
	File struct {
		io.ReadSeekCloser
		// ID file id.
		ID uuid.UUID
		// Size contains file size.
		Size int64
		// Metadata contains file meta info.
		Metadata json.RawMessage
	}
)
