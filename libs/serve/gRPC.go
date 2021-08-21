package serve

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	"github.com/Meat-Hook/back-template/libs/log"
)

// GRPC starts gRPC server on addr, logged as service.
// It runs until failed or ctx.Done.
func GRPC(logger zerolog.Logger, host string, port int, srv *grpc.Server) func(context.Context) error {
	return func(ctx context.Context) error {
		ln, err := net.Listen("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
		if err != nil {
			return fmt.Errorf("net.Listen: %w", err)
		}

		errc := make(chan error, 1)
		go func() { errc <- srv.Serve(ln) }()
		logger.Info().Str(log.Host, host).Int(log.Port, port).Msg("started")
		defer logger.Info().Msg("shutdown")

		select {
		case err = <-errc:
		case <-ctx.Done():
			srv.GracefulStop() // It will not interrupt streaming.
		}
		if err != nil {
			return fmt.Errorf("srv.Serve: %w", err)
		}

		return nil
	}
}
