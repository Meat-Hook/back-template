package app_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/gofrs/uuid"

	"github.com/Meat-Hook/back-template/cmd/file/internal/app"
)

func TestModule_UploadFile(t *testing.T) {
	t.Parallel()

	module, m, assert := start(t)

	var (
		file   = &bytes.Buffer{}
		fileID = uuid.Must(uuid.NewV4())
	)

	testCases := []struct {
		name    string
		file    io.Reader
		want    uuid.UUID
		wantErr error
	}{
		{"success", file, fileID, nil},
	}

	m.repo.EXPECT().Save(ctx, file).Return(fileID, nil)

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res, err := module.UploadFile(ctx, tc.file)
			assert.ErrorIs(err, tc.wantErr)
			assert.Equal(tc.want, res)
		})
	}
}

func TestModule_GetFile(t *testing.T) {
	t.Parallel()

	module, m, assert := start(t)

	var (
		fileID = uuid.Must(uuid.NewV4())
		file   = &app.File{
			ReadSeekCloser: nil,
			ID:             fileID,
			Size:           100,
			Metadata:       nil,
		}
	)

	testCases := []struct {
		name    string
		fileID  uuid.UUID
		want    *app.File
		wantErr error
	}{
		{"success", fileID, file, nil},
	}

	m.repo.EXPECT().Read(ctx, fileID).Return(file, nil)

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res, err := module.GetFile(ctx, tc.fileID)
			assert.ErrorIs(err, tc.wantErr)
			assert.Equal(tc.want, res)
		})
	}
}

func TestModule_SetMetadata(t *testing.T) {
	t.Parallel()

	module, m, assert := start(t)

	var (
		fileID   = uuid.Must(uuid.NewV4())
		metadata = json.RawMessage(`{ "field": "value" }`)
	)

	testCases := []struct {
		name     string
		fileID   uuid.UUID
		metadata json.RawMessage
		want     error
	}{
		{"success", fileID, metadata, nil},
	}

	m.repo.EXPECT().SetMetadata(ctx, fileID, metadata).Return(nil)

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := module.SetMetadata(ctx, tc.fileID, tc.metadata)
			assert.ErrorIs(err, tc.want)
		})
	}
}

func TestModule_Delete(t *testing.T) {
	t.Parallel()

	module, m, assert := start(t)

	var (
		fileID = uuid.Must(uuid.NewV4())
	)

	testCases := []struct {
		name   string
		fileID uuid.UUID
		want   error
	}{
		{"success", fileID, nil},
	}

	m.repo.EXPECT().Delete(ctx, fileID).Return(nil)

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := module.Delete(ctx, tc.fileID)
			assert.ErrorIs(err, tc.want)
		})
	}
}
