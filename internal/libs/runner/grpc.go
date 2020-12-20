package runner

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/Meat-Hook/back-template/internal/libs/log"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

// GRPC run grpc server.
func GRPC(logger zerolog.Logger, srv *grpc.Server, host string, port int) func(context.Context) error {
	return func(ctx context.Context) error {
		ln, err := net.Listen("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
		if err != nil {
			return fmt.Errorf("listen grpc: %w", err)
		}

		errc := make(chan error, 1)
		go func() { errc <- srv.Serve(ln) }()
		logger.Info().Str(log.Host, host).Int(log.Port, port).Msg("started")
		defer logger.Info().Msg("shutdown")

		select {
		case err = <-errc:
		case <-ctx.Done():
			srv.GracefulStop()
		}
		if err != nil {
			return fmt.Errorf("failed in grpc server: %w", err)
		}

		return nil
	}
}
