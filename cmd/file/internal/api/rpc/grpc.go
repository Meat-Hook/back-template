// Package rpc contains all methods for working grpc server.
package rpc

import (
	"context"
	"encoding/json"
	"io"

	"github.com/gofrs/uuid"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	"github.com/Meat-Hook/back-template/libs/rpc"
	pb "github.com/Meat-Hook/back-template/proto/gen/go/file/v1"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

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
func New(ctx context.Context, applications files, metric *grpc_prometheus.ServerMetrics) *grpc.Server {
	logger := zerolog.Ctx(ctx)
	srv := rpc.Server(*logger, metric)
	pb.RegisterServiceServer(srv, &api{app: applications})

	return srv
}
