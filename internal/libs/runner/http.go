package runner

import (
	"context"
	"fmt"
	"net"

	"github.com/Meat-Hook/back-template/internal/libs/log"
	"github.com/rs/zerolog"
)

type swagger interface {
	Serve() error
	Shutdown() error
}

// HTTP run http server.
func HTTP(ctx context.Context, logger zerolog.Logger, srv swagger, ip net.IP, port int) func() error {
	return func() error {
		errc := make(chan error, 1)
		go func() { errc <- srv.Serve() }()
		logger.Info().IPAddr(log.Host, ip).Int(log.Port, port).Msg("started")
		defer logger.Info().Msg("shutdown")

		select {
		case err := <-errc:
			if err != nil {
				return fmt.Errorf("failed to listen http server: %w", err)
			}
		case <-ctx.Done():
			err := srv.Shutdown()
			if err != nil {
				return fmt.Errorf("failed to shutdown http server: %w", err)
			}
		}

		return nil
	}
}
