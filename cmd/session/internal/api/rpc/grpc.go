// Package rpc contains all methods for working grpc server.
package rpc

import (
	"context"

	"github.com/gofrs/uuid"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	"github.com/Meat-Hook/back-template/cmd/session/internal/app"
	"github.com/Meat-Hook/back-template/libs/rpc"
	pb "github.com/Meat-Hook/back-template/proto/gen/go/session/v1"
)

// Metric contains general metrics for gRPC methods.
var metric struct { //nolint:gochecknoglobals // Metrics are global anyway.
	server *grpc_prometheus.ServerMetrics
}

// For convenient testing.
// Wrapper for app.Module.
type sessions interface {
	Session(ctx context.Context, token string) (*app.Session, error)
	NewSession(ctx context.Context, userID uuid.UUID, origin app.Origin) (*app.Token, error)
	RemoveSession(ctx context.Context, sessionID uuid.UUID) error
}

type api struct {
	app sessions
}

// New creates and returns gRPC server.
func New(ctx context.Context, req *prometheus.Registry, namespace string, applications sessions) *grpc.Server {
	logger := zerolog.Ctx(ctx)
	metric.server = rpc.NewServerMetrics(req, namespace)

	srv := rpc.Server(*logger, metric.server)
	pb.RegisterServiceServer(srv, &api{app: applications})

	return srv
}
