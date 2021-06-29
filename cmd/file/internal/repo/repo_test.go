package repo_test

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/Meat-Hook/back-template/cmd/file/internal/app"
	"github.com/Meat-Hook/back-template/cmd/file/internal/repo"
	"github.com/Meat-Hook/back-template/libs/metrics"
)

func TestRepo_Smoke(t *testing.T) {
	t.Parallel()

	db, assert := start(t)

	m := metrics.DB("file", metrics.MethodsOf(&repo.Repo{})...)
	r := repo.New(db, &m)

	f, err := os.Open(testFile)
	assert.Nil(err)
	defer func() {
		assert.Nil(f.Close())
	}()

	fileID, err := r.Save(ctx, f)
	assert.Nil(err)
	assert.NotNil(fileID)

	_, err = f.Seek(0, io.SeekStart)
	assert.Nil(err)

	fFromDB, err := r.Read(ctx, fileID)
	assert.Nil(err)

	bufOldFile, err := io.ReadAll(f)
	assert.Nil(err)

	bufNewFile, err := io.ReadAll(fFromDB)
	assert.Nil(err)

	assert.Equal(bufOldFile, bufNewFile)

	fFromDB.Metadata = json.RawMessage(`{"hello": "world!"}`)
	err = r.SetMetadata(ctx, fFromDB.ID, fFromDB.Metadata)
	assert.Nil(err)

	updatedFileFromDB, err := r.Read(ctx, fFromDB.ID)
	assert.Nil(err)

	assert.Equal(fFromDB.Metadata, updatedFileFromDB.Metadata)

	err = r.Delete(ctx, fFromDB.ID)
	assert.Nil(err)

	newF, err := r.Read(ctx, fFromDB.ID)
	assert.Nil(newF)
	assert.ErrorIs(err, app.ErrNotFound)
}
