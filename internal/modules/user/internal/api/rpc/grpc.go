// Package rpc contains all methods for working grpc server.
package rpc

import (
	"context"

	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/rpc/pb"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/app"
	prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

//go:generate mockgen -source=grpc.go -destination mock.app.contracts_test.go -package rpc_test

// For convenient testing.
type users interface {
	// Access check user access by email and pass and returns user info.
	Access(ctx context.Context, email, password string) (*app.User, error)
}

type api struct {
	app users
}

// New register service by grpc.Server and register metrics.
func New(app users, srv *grpc.Server) *grpc.Server {
	pb.RegisterUserServer(srv, &api{app: app})

	prometheus.Register(srv)

	return srv
}
