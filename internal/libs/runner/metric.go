package runner

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/Meat-Hook/back-template/internal/libs/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

// Metric run metric for collect service metric.
func Metric(ctx context.Context, logger zerolog.Logger, ip net.IP, port int) func() error {
	return func() error {
		http.Handle("/metrics", promhttp.Handler())
		srv := &http.Server{
			Addr: net.JoinHostPort(ip.String(), strconv.Itoa(port)),
		}

		errc := make(chan error, 1)
		go func() { errc <- srv.ListenAndServe() }()
		logger.Info().IPAddr(log.Host, ip).Int(log.Port, port).Msg("started")
		defer logger.Info().Msg("shutdown")

		select {
		case err := <-errc:
			if err != nil {
				return fmt.Errorf("failed to listen http server: %w", err)
			}
		case <-ctx.Done():
			ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
			defer cancel()

			err := srv.Shutdown(ctx)
			if err != nil {
				return fmt.Errorf("failed to shutdown http server: %w", err)
			}
		}

		return nil
	}
}
