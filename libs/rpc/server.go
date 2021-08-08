package rpc

import (
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

// Server returns gRPC server configured to listen on the TCP network.
func Server(
	logger zerolog.Logger,
	serverMetrics *grpc_prometheus.ServerMetrics,
) *grpc.Server {
	srv := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    keepaliveTime,
			Timeout: keepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             keepaliveMinTime,
			PermitWithoutStream: true,
		}),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			serverMetrics.UnaryServerInterceptor(),
			MakeUnaryServerLogger(logger),
			MakeUnaryServerRecover(),
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandlerContext(recoveryFunc)),
			UnaryServerAccessLog,
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			prometheus.StreamServerInterceptor,
			MakeStreamServerLogger(logger),
			MakeStreamServerRecover(),
			grpc_recovery.StreamServerInterceptor(grpc_recovery.WithRecoveryHandlerContext(recoveryFunc)),
			StreamServerAccessLog,
		)),
	)

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthServer)

	return srv
}
