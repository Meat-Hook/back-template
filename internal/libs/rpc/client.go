package rpc

import (
	"context"
	"fmt"
	"time"

	"github.com/Meat-Hook/back-template/internal/libs/middleware"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// Client build and returns new grpc client conn.
func Client(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                50 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			prometheus.UnaryClientInterceptor,
			middleware.UnaryClientLogger,
			middleware.UnaryClientAccessLog,
		)),
		grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
			prometheus.StreamClientInterceptor,
			middleware.StreamClientLogger,
			middleware.StreamClientAccessLog,
		)),
		grpc.WithReadBufferSize(68*1024),
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("grpc dial: %w", err)
	}

	return conn, nil
}
