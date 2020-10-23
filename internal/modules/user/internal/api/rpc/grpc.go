// Package rpc contains all methods for working grpc server.
package rpc

import (
	"context"
	"time"

	"github.com/Meat-Hook/back-template/internal/libs/middleware"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/rpc/pb"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/app"
	groc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

//go:generate mockgen -source=grpc.go -destination mock.app.contracts_test.go -package rpc_test

// For convenient testing.
type users interface {
	// UserByID get user by id.
	UserByID(ctx context.Context, session app.Session, id int) (*app.User, error)
}

type api struct {
	app users
}

// New returns gRPC server configured to listen on the TCP network.
func New(app users, logger zerolog.Logger) *grpc.Server {
	server := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    50 * time.Second,
			Timeout: 10 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             30 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.UnaryInterceptor(groc_middleware.ChainUnaryServer(
			prometheus.UnaryServerInterceptor,
			middleware.MakeUnaryServerLogger(logger.With()),
			middleware.UnaryServerRecover,
			middleware.UnaryServerAccessLog,
		)),
	)

	pb.RegisterUserServer(server, &api{app: app})

	prometheus.Register(server)

	return server
}
