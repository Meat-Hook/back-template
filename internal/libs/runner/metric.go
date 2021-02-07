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
func Metric(logger zerolog.Logger, host string, port int) func(context.Context) error {
	return func(ctx context.Context) error {
		handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer logger.Debug().Msg("collect metrics")
			promhttp.Handler().ServeHTTP(w, r)
		}))
		http.Handle("/metrics", handler)
		srv := &http.Server{
			Addr: net.JoinHostPort(host, strconv.Itoa(port)),
		}

		errc := make(chan error, 1)
		go func() { errc <- srv.ListenAndServe() }()
		logger.Info().Str(log.Host, host).Int(log.Port, port).Msg("started")
		defer logger.Info().Msg("shutdown")

		select {
		case err := <-errc:
			if err != nil {
				return fmt.Errorf("failed to listen http server: %w", err)
			}
		case <-ctx.Done():
			ctxDone, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
			defer cancel()

			err := srv.Shutdown(ctxDone)
			if err != nil {
				return fmt.Errorf("failed to shutdown http server: %w", err)
			}
		}

		return nil
	}
}
