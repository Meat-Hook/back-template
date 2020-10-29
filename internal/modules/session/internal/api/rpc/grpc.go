// Package rpc contains all methods for working grpc server.
package rpc

import (
	"context"

	"github.com/Meat-Hook/back-template/internal/modules/session/internal/api/rpc/pb"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/app"
	prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

//go:generate mockgen -source=grpc.go -destination mock.app.contracts_test.go -package rpc_test

// For convenient testing.
type sessions interface {
	// Session get user session by his token.
	Session(ctx context.Context, token string) (*app.Session, error)
}

type api struct {
	app sessions
}

// New register service by grpc.Server and register metrics.
func New(app sessions, srv *grpc.Server) *grpc.Server {
	pb.RegisterSessionServer(srv, &api{app: app})

	prometheus.Register(srv)

	return srv
}
