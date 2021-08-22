// Package repo contains implements for app.Repo.
// Provide file chunk to and from repository.
package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgtype"
	"github.com/jmoiron/sqlx"

	"github.com/Meat-Hook/back-template/cmd/file/internal/app"
	"github.com/Meat-Hook/back-template/libs/db"
)

var _ app.Repo = &Repo{}

type (
	// Repo provided data from and to database.
	Repo struct {
		db *db.DB
	}

	fileInfo struct {
		ID        pgtype.UUID      `db:"id"`
		Size      int64            `db:"size"`
		Metadata  pgtype.JSONB     `db:"metadata"`
		ChunkIDs  pgtype.UUIDArray `db:"chunk_ids"`
		CreatedAt pgtype.Timestamp `db:"created_at"`
		UpdatedAt pgtype.Timestamp `db:"updated_at"`
	}

	chunk struct {
		ID        pgtype.UUID      `db:"id"`
		FileID    pgtype.UUID      `db:"file_id"`
		Bytes     pgtype.Bytea     `db:"bytes"`
		CreatedAt pgtype.Timestamp `db:"created_at"`
		UpdatedAt pgtype.Timestamp `db:"updated_at"`
	}
)

func (f *fileInfo) convert(r *file) *app.File {
	return &app.File{
		ReadSeekCloser: r,
		ID:             f.ID.Bytes,
		Size:           f.Size,
		Metadata:       f.Metadata.Bytes,
	}
}

// New build and returns user db.
func New(r *db.DB) *Repo {
	return &Repo{
		db: r,
	}
}

// Save for implements app.Repo.
func (r *Repo) Save(ctx context.Context, reader io.Reader) (res uuid.UUID, err error) {
	err = r.db.Tx(ctx, nil, func(tx *sqlx.Tx) (err error) {
		const querySaveFile = `
			insert into files default values returning *
		`

		file := &fileInfo{}
		err = tx.QueryRowxContext(ctx, querySaveFile).StructScan(file)
		if err != nil {
			return fmt.Errorf("tx.QueryRowxContext: %w", convertErr(err))
		}

		const querySaveChunk = `insert into chunks (file_id, bytes) values ($1, $2) returning *`

		buf := make([]byte, app.MaxChunkSize)
		i := 0
		for {
			n, err := reader.Read(buf)
			if err != nil && !errors.Is(err, io.EOF) {
				return fmt.Errorf("reader.Read: %w", convertErr(err))
			}

			if n == 0 {
				break
			}

			fileChunk := &chunk{}
			err = tx.QueryRowxContext(ctx, querySaveChunk, file.ID, pgtype.Bytea{
				Bytes:  buf[:n],
				Status: pgtype.Present,
			}).StructScan(fileChunk)
			if err != nil {
				return fmt.Errorf("tx.QueryRowxContext: %w", convertErr(err))
			}

			file.ChunkIDs.Elements = append(file.ChunkIDs.Elements, fileChunk.ID)
			file.Size += int64(n)
			i++
		}

		const queryUpdateSizeAndChunks = `
			update files set size = $1, chunk_ids = $2 where id = $3;
		`

		file.ChunkIDs.Status = pgtype.Present
		file.ChunkIDs.Dimensions = append(file.ChunkIDs.Dimensions,
			pgtype.ArrayDimension{
				Length:     int32(len(file.ChunkIDs.Elements)),
				LowerBound: 1,
			})
		_, err = tx.ExecContext(ctx, queryUpdateSizeAndChunks, file.Size, file.ChunkIDs, file.ID)
		if err != nil {
			return fmt.Errorf("tx.ExecContext: %w", convertErr(err))
		}

		res = file.ID.Bytes

		return nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	return res, nil
}

// Read for implements app.Repo.
func (r *Repo) Read(ctx context.Context, fileID uuid.UUID) (res *app.File, err error) {
	err = r.db.NoTx(func(db *sqlx.DB) error {
		const query = `select * from files where id = $1;`
		fInfo := &fileInfo{}

		err := db.GetContext(ctx, fInfo, query, fileID)
		if err != nil {
			return fmt.Errorf("db.GetContext: %w", convertErr(err))
		}

		f := &file{
			db:          db,
			chunks:      fInfo.ChunkIDs,
			isClosed:    false,
			size:        fInfo.Size,
			position:    0,
			chunkCached: -1,
			chunkCache:  make([]byte, app.MaxChunkSize),
			error:       nil,
		}

		res = fInfo.convert(f)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// SetMetadata for implements app.Repo.
func (r *Repo) SetMetadata(ctx context.Context, fileID uuid.UUID, metadata json.RawMessage) error {
	return r.db.NoTx(func(db *sqlx.DB) error {
		const query = `update files set metadata = $1 where id = $2`

		convertMetadata := pgtype.JSONB{
			Bytes:  metadata,
			Status: pgtype.Present,
		}

		result, err := db.ExecContext(ctx, query, convertMetadata, fileID)
		if err != nil {
			return fmt.Errorf("db.ExecContext: %w", convertErr(err))
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("result.RowsAffected: %w", err)
		}

		if rowsAffected == 0 {
			return app.ErrNotFound
		}

		return nil
	})
}

// Delete for implements app.Repo.
func (r *Repo) Delete(ctx context.Context, fileID uuid.UUID) error {
	return r.db.NoTx(func(db *sqlx.DB) error {
		const query = `
		delete
		from files
		where id = $1`

		_, err := db.ExecContext(ctx, query, fileID)
		if err != nil {
			return fmt.Errorf("db.ExecContext: %w", convertErr(err))
		}

		return nil
	})
}
