package rpc

import (
	"errors"
	"fmt"
	"io"

	pb "github.com/Meat-Hook/back-template/proto/gen/go/file/v1"
)

var _ io.Reader = &Reader{}

// Reader implements io.Reader for gRPC stream.
type Reader struct {
	stream pb.FileService_UploadServer
	cache  []byte
}

// NewReader build new instance gRPC reader.
func NewReader(stream pb.FileService_UploadServer) io.Reader {
	return &Reader{
		stream: stream,
		cache:  []byte{},
	}
}

// Read for implemented io.Reader.
func (r *Reader) Read(b []byte) (sent int, err error) {
	var res []byte

	if r.cacheIsEmpty() {
		msg, err := r.stream.Recv()
		if err != nil && !errors.Is(err, io.EOF) {
			return 0, fmt.Errorf("stream recv: %w", err)
		}

		if errors.Is(err, io.EOF) {
			return 0, io.EOF
		}

		res = msg.Chunk.Content
	} else {
		res = r.pullCache()
	}

	if len(res) == 0 {
		return 0, io.EOF
	}

	sent = copy(b, res)
	if sent < len(res) {
		r.cache = res[sent:]
	}

	return sent, nil
}

func (r *Reader) cacheIsEmpty() bool {
	return len(r.cache) == 0
}

func (r *Reader) pullCache() []byte {
	res := r.cache
	r.cache = []byte{}

	return res
}
