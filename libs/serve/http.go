package serve

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"

	"github.com/Meat-Hook/back-template/libs/log"
)

// HTTP starts HTTP server on addr using handler logged as service.
// It runs until failed or ctx.Done.
func HTTP(logger zerolog.Logger, host string, port int, handler http.Handler) func(context.Context) error {
	return func(ctx context.Context) error {
		srv := &http.Server{
			Addr:    net.JoinHostPort(host, strconv.Itoa(port)),
			Handler: handler,
		}

		errc := make(chan error, 1)
		go func() { errc <- srv.ListenAndServe() }()
		logger.Info().Str(log.Host, host).Int(log.Port, port).Msg("started")
		defer logger.Info().Msg("shutdown")

		var err error
		select {
		case err = <-errc:
		case <-ctx.Done():
			err = srv.Shutdown(context.Background())
		}
		if err != nil {
			return fmt.Errorf("srv.ListenAndServe: %w", err)
		}

		return nil
	}
}

// Metrics starts HTTP server on addr path /metrics using reg as
// prometheus handler.
func Metrics(logger zerolog.Logger, host string, port int, reg *prometheus.Registry) func(context.Context) error {
	return func(ctx context.Context) error {
		mux := http.NewServeMux()
		HandleMetrics(mux, reg)

		return HTTP(logger, host, port, mux)(ctx)
	}
}

// HandleMetrics adds reg's prometheus handler on /metrics at mux.
func HandleMetrics(mux *http.ServeMux, reg *prometheus.Registry) {
	handler := promhttp.InstrumentMetricHandler(reg, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	mux.Handle("/metrics", handler)
}
