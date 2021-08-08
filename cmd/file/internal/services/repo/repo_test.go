package repo_test

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/Meat-Hook/back-template/cmd/file/internal/app"
)

func TestRepo_Smoke(t *testing.T) {
	t.Parallel()

	ctx, r, assert := start(t)

	f, err := os.Open(testFile)
	assert.NoError(err)
	defer func() {
		assert.NoError(f.Close())
	}()

	fileID, err := r.Save(ctx, f)
	assert.NoError(err)
	assert.NotNil(fileID)

	_, err = f.Seek(0, io.SeekStart)
	assert.NoError(err)

	fFromDB, err := r.Read(ctx, fileID)
	assert.NoError(err, context.Canceled)

	bufOldFile, err := io.ReadAll(f)
	assert.NoError(err)

	bufNewFile, err := io.ReadAll(fFromDB)
	assert.NoError(err)

	assert.Equal(bufOldFile, bufNewFile)

	fFromDB.Metadata = json.RawMessage(`{"hello": "world!"}`)
	err = r.SetMetadata(ctx, fFromDB.ID, fFromDB.Metadata)
	assert.NoError(err)

	updatedFileFromDB, err := r.Read(ctx, fFromDB.ID)
	assert.NoError(err)

	assert.Equal(fFromDB.Metadata, updatedFileFromDB.Metadata)

	err = r.Delete(ctx, fFromDB.ID)
	assert.NoError(err)

	newF, err := r.Read(ctx, fFromDB.ID)
	assert.Nil(newF)
	assert.ErrorIs(err, app.ErrNotFound)
}
