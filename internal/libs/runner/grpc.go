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

func GRPC(ctx context.Context, logger zerolog.Logger, srv *grpc.Server, ip net.IP, port int) func() error {
	return func() error {
		ln, err := net.Listen("tcp", net.JoinHostPort(ip.String(), strconv.Itoa(port)))
		if err != nil {
			return fmt.Errorf("listen grpc: %w", err)
		}

		errc := make(chan error, 1)
		go func() { errc <- srv.Serve(ln) }()
		logger.Info().IPAddr(log.Host, ip).Int(log.Port, port).Msg("started")
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
