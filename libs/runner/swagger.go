package runner

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	"github.com/Meat-Hook/back-template/libs/log"
)

type swaggerServer interface {
	Serve() error
	Shutdown() error
}

// Swagger run swagger web server.
func Swagger(logger zerolog.Logger, srv swaggerServer, host string, port int) func(context.Context) error {
	return func(ctx context.Context) error {
		errc := make(chan error, 1)
		go func() { errc <- srv.Serve() }()
		logger.Info().Str(log.Host, host).Int(log.Port, port).Msg("started")
		defer logger.Info().Msg("shutdown")

		select {
		case err := <-errc:
			if err != nil {
				return fmt.Errorf("failed to listen web server: %w", err)
			}
		case <-ctx.Done():
			err := srv.Shutdown()
			if err != nil {
				return fmt.Errorf("failed to shutdown web server: %w", err)
			}
		}

		return nil
	}
}
