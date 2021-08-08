package repo

import (
	"fmt"
	"io"
	"os"

	"github.com/jackc/pgtype"
	"github.com/jmoiron/sqlx"

	"github.com/Meat-Hook/back-template/cmd/file/internal/app"
)

var _ io.ReadSeekCloser = &file{}

type file struct {
	db          *sqlx.DB
	chunks      pgtype.UUIDArray
	isClosed    bool
	size        int64
	position    int64
	chunkCached int64
	chunkCache  []byte
	error       error
}

// Read for implemented io.Reader.
func (f *file) Read(dst []byte) (n int, err error) {
	switch {
	case f.error != nil:
		return 0, f.error
	case f.isClosed:
		return 0, os.ErrClosed
	case f.position >= f.size:
		return 0, io.EOF
	case f.position < 0:
		return 0, ErrNegativePosition
	}

	stoppedAtChunk := int(f.position / app.MaxChunkSize)
	stoppedAtIndex := int(f.position % app.MaxChunkSize)

	chunk := &chunk{}
	if stoppedAtChunk != int(f.chunkCached) {
		const query = `select * from chunks where id = $1;`
		err = f.db.Get(chunk, query, f.chunks.Elements[stoppedAtChunk])
		if err != nil {
			return 0, f.lastErr(fmt.Errorf("f.db.Get: %w", convertErr(err)))
		}

		f.chunkCached = int64(stoppedAtChunk)
		f.chunkCache = chunk.Bytes.Bytes
	}

	n = copy(dst, f.chunkCache[stoppedAtIndex:])
	f.position += int64(n)
	if f.position >= f.size {
		return n, f.lastErr(io.EOF)
	}

	return n, nil
}

// Seek for implemented io.Seeker.
func (f *file) Seek(offset int64, whence int) (int64, error) {
	switch {
	case f.error != nil:
		return 0, f.error
	case f.isClosed:
		return 0, os.ErrClosed
	}

	newPosition := f.position
	switch whence {
	case io.SeekStart:
		newPosition = offset
	case io.SeekCurrent:
		newPosition += offset
	case io.SeekEnd:
		newPosition = f.size + offset
	default:
		return 0, ErrUnexpectedWhence
	}

	if newPosition < 0 {
		return 0, ErrNegativePosition
	}

	f.position = newPosition

	return f.position, nil
}

// Close for implemented io.Closer.
func (f *file) Close() error {
	if f.isClosed {
		return f.lastErr(os.ErrClosed)
	}

	f.isClosed = true

	return nil
}

func (f *file) lastErr(err error) error {
	f.error = err

	return f.error
}
