// Package rpc contains all methods for working grpc server.
package rpc

import (
	"context"

	app2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/app"
	pb "github.com/Meat-Hook/back-template/proto/gen/go/session/v1"
	prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

//go:generate mockgen -source=grpc.go -destination mock.app.contracts_test.go -package rpc_test

// For convenient testing.
type sessions interface {
	// Session get user session by his token.
	Session(ctx context.Context, token string) (*app2.Session, error)
}

type api struct {
	app sessions
}

// New register service by grpc.Server and register metrics.
func New(applications sessions, srv *grpc.Server) *grpc.Server {
	pb.RegisterSessionServiceServer(srv, &api{app: applications})

	prometheus.Register(srv)

	return srv
}
