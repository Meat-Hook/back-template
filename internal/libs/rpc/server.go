package rpc

import (
	"time"

	"github.com/Meat-Hook/back-template/internal/libs/middleware"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// Server returns gRPC server configured to listen on the TCP network.
func Server(logger zerolog.Logger) *grpc.Server {
	return grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    50 * time.Second,
			Timeout: 10 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             30 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			prometheus.UnaryServerInterceptor,
			middleware.MakeUnaryServerLogger(logger.With()),
			middleware.UnaryServerRecover,
			middleware.UnaryServerAccessLog,
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			prometheus.StreamServerInterceptor,
			middleware.MakeStreamServerLogger(logger.With()),
			middleware.StreamServerRecover,
			middleware.StreamServerAccessLog,
		)),
	)
}
