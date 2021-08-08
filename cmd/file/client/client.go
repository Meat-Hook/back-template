// Package client provide to internal method of service file.
package client

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/Meat-Hook/back-template/cmd/file/internal/app"
	"github.com/Meat-Hook/back-template/libs/log"
	pb "github.com/Meat-Hook/back-template/proto/gen/go/file/v1"
)

// Client to file microservice.
type Client struct {
	conn pb.ServiceClient
}

// New build and returns new client to microservice session.
func New(conn grpc.ClientConnInterface) *Client {
	return &Client{conn: pb.NewServiceClient(conn)}
}

// Errors.
var (
	ErrNotFound = app.ErrNotFound
)

// Upload file to database.
func (c *Client) Upload(ctx context.Context, r io.Reader) (uuid.UUID, error) {
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		log.ReqID: []string{log.ReqIDFromCtx(ctx)},
	})

	stream, err := c.conn.Upload(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("c.conn.Upload: %w", err)
	}

	buf := make([]byte, app.MaxChunkSize)

	for {
		n, err := r.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			return uuid.Nil, fmt.Errorf("r.Read: %w", err)
		}

		if n == 0 {
			break
		}

		in := &pb.UploadRequest{
			Chunk: &pb.Chunk{
				Content: buf[:n],
			},
		}

		err = stream.Send(in)
		if err != nil {
			return uuid.Nil, fmt.Errorf("stream.Send: %w", err)
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		return uuid.Nil, fmt.Errorf("stream.CloseAndRecv: %w", err)
	}

	id, err := uuid.FromString(res.FileId.Value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("uuid.FromString: %w", err)
	}

	return id, nil
}

// SetMetadata set file metadata.
func (c *Client) SetMetadata(ctx context.Context, fileID uuid.UUID, fileMD map[string]interface{}) error {
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		log.ReqID: []string{log.ReqIDFromCtx(ctx)},
	})

	details, err := structpb.NewStruct(fileMD)
	if err != nil {
		return fmt.Errorf("structpb.NewStruct: %w", err)
	}

	in := &pb.SetMetadataRequest{
		FileId:   &pb.UUID{Value: fileID.String()},
		Metadata: &pb.Metadata{Details: details},
	}

	_, err = c.conn.SetMetadata(ctx, in)
	switch {
	case status.Code(err) == codes.NotFound:
		return ErrNotFound
	case err != nil:
		return fmt.Errorf("c.conn.SetMetadata: %w", err)
	}

	return nil
}

// Delete file from database.
func (c *Client) Delete(ctx context.Context, fileID uuid.UUID) error {
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		log.ReqID: []string{log.ReqIDFromCtx(ctx)},
	})

	in := &pb.DeleteRequest{
		FileId: &pb.UUID{
			Value: fileID.String(),
		},
	}

	_, err := c.conn.Delete(ctx, in)
	if err != nil {
		return fmt.Errorf("c.conn.Delete: %w", err)
	}

	return nil
}
