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

	testCases := map[string]struct {
		filePath string
		appErr   error
		wantErr  *models.Error
	}{
		"success": {testFile, nil, nil},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			file, err := os.Open(tc.filePath)
			assert.NoError(err)
			stat, err := file.Stat()
			assert.NoError(err)

			fileBuf, err := io.ReadAll(file)
			assert.NoError(err)
			_, err = file.Seek(0, io.SeekStart)
			assert.NoError(err)

			appFile := &app.File{
				ReadSeekCloser: file,
				ID:             uuid.Must(uuid.NewV4()),
				Size:           stat.Size(),
				Metadata:       nil,
			}

			_, mockApp, client, assert := start(t)

			mockApp.EXPECT().GetFile(gomock.Any(), appFile.ID).Return(appFile, tc.appErr)

			b := &bytes.Buffer{}
			params := operations.NewGetFileParams().
				WithID(strfmt.UUID(appFile.ID.String()))

			res, err := client.Operations.GetFile(params, b)
			if tc.wantErr == nil {
				assert.NoError(err)
				assert.Equal(fileBuf, b.Bytes())
			} else {
				assert.NoError(res)
				assert.Equal(tc.wantErr, errPayload(err))
			}
		})
	}
}
