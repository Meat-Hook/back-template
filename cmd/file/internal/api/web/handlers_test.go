package web_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/Meat-Hook/back-template/cmd/file/internal/api/web/generated/client/operations"
	"github.com/Meat-Hook/back-template/cmd/file/internal/api/web/generated/models"
	"github.com/Meat-Hook/back-template/cmd/file/internal/app"
)

func TestService_GetFile(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	file, err := os.Open(testFile)
	assert.NoError(err)
	stat, err := file.Stat()
	assert.NoError(err)

	fileBuf, err := io.ReadAll(file)
	assert.NoError(err)
	_, err = file.Seek(0, io.SeekStart)
	assert.NoError(err)

	appFile := app.File{
		ReadSeekCloser: file,
		ID:             uuid.Must(uuid.NewV4()),
		Size:           stat.Size(),
		Metadata:       nil,
	}

	testCases := map[string]struct {
		fileID  uuid.UUID
		appRes  *app.File
		appErr  error
		want    []byte
		wantErr *models.Error
	}{
		"success": {appFile.ID, &appFile, nil, fileBuf, nil},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, mockApp, client, assert := start(t)

			mockApp.EXPECT().GetFile(gomock.Any(), tc.fileID).Return(tc.appRes, tc.appErr)

			b := &bytes.Buffer{}
			params := operations.NewGetFileParams().
				WithID(strfmt.UUID(tc.fileID.String()))

			res, err := client.Operations.GetFile(params, b)
			if tc.wantErr == nil {
				assert.NoError(err)
				assert.Equal(tc.want, b.Bytes())
			} else {
				assert.NoError(res)
				assert.Equal(tc.wantErr, errPayload(err))
			}
		})
	}
}
