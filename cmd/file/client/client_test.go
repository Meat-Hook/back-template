package client_test

import (
	"encoding/json"
	"testing"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Meat-Hook/back-template/cmd/file/client"
	"github.com/Meat-Hook/back-template/cmd/file/internal/app"
)

func TestClient_Delete(t *testing.T) {
	t.Parallel()

	fileID := uuid.Must(uuid.NewV4())
	conn, _, assert := start(t, fileID, nil, nil)

	testCases := map[string]struct {
		fileID uuid.UUID
		want   error
	}{
		"success":       {fileID, nil},
		"err_not_found": {uuid.Must(uuid.NewV4()), status.Error(codes.NotFound, app.ErrNotFound.Error())},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			err := conn.Delete(ctx, tc.fileID)
			assert.ErrorIs(err, tc.want)
		})
	}
}

func TestClient_SetMetadata(t *testing.T) {
	t.Parallel()

	fileID := uuid.Must(uuid.NewV4())
	md := json.RawMessage(`{"field":"value"}`)
	conn, _, assert := start(t, fileID, md, nil)

	testCases := map[string]struct {
		fileID uuid.UUID
		md     map[string]interface{}
		want   error
	}{
		"success":       {fileID, map[string]interface{}{"field": "value"}, nil},
		"err_not_found": {uuid.Must(uuid.NewV4()), nil, client.ErrNotFound},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			err := conn.SetMetadata(ctx, tc.fileID, tc.md)
			assert.ErrorIs(err, tc.want)
		})
	}
}
