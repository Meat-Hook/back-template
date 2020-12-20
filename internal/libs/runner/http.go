package runner

import (
	"context"
	"fmt"

	"github.com/Meat-Hook/back-template/internal/libs/log"
	"github.com/rs/zerolog"
)

type swagger interface {
	Serve() error
	Shutdown() error
}

// HTTP run http server.
func HTTP(logger zerolog.Logger, srv swagger, host string, port int) func(context.Context) error {
	return func(ctx context.Context) error {
		errc := make(chan error, 1)
		go func() { errc <- srv.Serve() }()
		logger.Info().Str(log.Host, host).Int(log.Port, port).Msg("started")
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
