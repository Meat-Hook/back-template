package rpc

import (
	"context"
	"fmt"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// Dial creates a gRPC client connection to the given target.
func Dial(ctx context.Context, logger zerolog.Logger, addr string, metrics *grpc_prometheus.ClientMetrics) (*grpc.ClientConn, error) {
	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                50 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			metrics.UnaryClientInterceptor(),
			MakeUnaryClientLogger(logger),
			UnaryClientAccessLog,
		)),
		grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
			metrics.StreamClientInterceptor(),
			MakeStreamClientLogger(logger),
			StreamClientAccessLog,
		)),
		grpc.WithReadBufferSize(68*1024),
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("grpc dial: %w", err)
	}

	return conn, nil
}
