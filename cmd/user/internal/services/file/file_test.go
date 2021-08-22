package file_test

import (
	"bytes"
	"testing"

	"github.com/gofrs/uuid"
)

func TestClient_Upload(t *testing.T) {
	t.Parallel()

	file := bytes.NewBuffer(uuid.Must(uuid.NewV4()).Bytes())

	testCases := []struct {
		name    string
		want    uuid.UUID
		wantErr error
	}{
		{"success", uuid.Must(uuid.NewV4()), nil},
		{"err_any", uuid.Nil, errAny},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc, mock, assert := start(t)

			mock.EXPECT().Upload(ctx, file).Return(tc.want, tc.wantErr)

			res, err := svc.Upload(ctx, file)
			assert.Equal(tc.want, res)
			assert.ErrorIs(err, tc.wantErr)
		})
	}
}

func TestClient_Delete(t *testing.T) {
	t.Parallel()

	fileID := uuid.Must(uuid.NewV4())

	testCases := []struct {
		name string
		want error
	}{
		{"success", nil},
		{"err_any", errAny},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc, mock, assert := start(t)

			mock.EXPECT().Delete(ctx, fileID).Return(tc.want)

			err := svc.Delete(ctx, fileID)
			assert.ErrorIs(err, tc.want)
		})
	}
}
