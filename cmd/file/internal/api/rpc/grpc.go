// Package rpc contains all methods for working grpc server.
package rpc

import (
	"context"
	"encoding/json"
	"io"

	pb "github.com/Meat-Hook/back-template/proto/gen/go/file/v1"
	"github.com/gofrs/uuid"
	prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

//go:generate mockgen -source=grpc.go -destination mock.app.contracts_test.go -package rpc_test

// For convenient testing.
// Wrapper for app.Module.
type files interface {
	UploadFile(ctx context.Context, file io.Reader) (uuid.UUID, error)
	SetMetadata(ctx context.Context, fileID uuid.UUID, metadata json.RawMessage) error
	Delete(ctx context.Context, fileID uuid.UUID) error
}

type api struct {
	app files
}

// New register service by grpc.Server and register metrics.
func New(applications files, srv *grpc.Server) *grpc.Server {
	pb.RegisterFileServiceServer(srv, &api{app: applications})

	prometheus.Register(srv)

	return srv
}
