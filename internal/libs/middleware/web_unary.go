package middleware

import (
	"net"
	"net/http"
	"strconv"

	"github.com/Meat-Hook/back-template/internal/libs/log"
	"github.com/Meat-Hook/back-template/internal/libs/metrics"
	"github.com/felixge/httpsnoop"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

// go-swagger responders panic on error while writing response to client,
// this shouldn't result in crash - unlike a real, reasonable panic.
//
// Usually it should be second middlewareFunc (after CreateLogger).
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				metrics.PanicsTotal.Inc()
				logger := zerolog.Ctx(r.Context())
				logger.Error().Msgf("panic with error: %v", err)

				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// CreateLogger create new logger by base path and zerolog builder.
func CreateLogger(builder zerolog.Context) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)

			newLogger := builder.
				IPAddr(log.IP, net.ParseIP(ip)).
				Str(log.HTTPMethod, r.Method).
				Str(log.Func, r.URL.Path).
				Stringer(log.Request, xid.New()).
				Logger()

			r = r.WithContext(newLogger.WithContext(r.Context()))

			next.ServeHTTP(w, r)
		})
	}
}

// AccessLog logs handled request.
func AccessLog(metric *metrics.API) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m := httpsnoop.CaptureMetrics(next, w, r)

			metric.ReqInFlight.Inc()
			defer metric.ReqInFlight.Dec()

			l := prometheus.Labels{
				metrics.ResourceLabel: r.URL.Path,
				metrics.MethodLabel:   r.Method,
				metrics.CodeLabel:     strconv.Itoa(m.Code),
			}
			metric.ReqTotal.With(l).Inc()
			metric.ReqDuration.With(l).Observe(m.Duration.Seconds())

			logger := zerolog.Ctx(r.Context())
			if m.Code < http.StatusInternalServerError {
				logger.Info().Int(log.Code, m.Code).Stringer(log.Duration, m.Duration).Msg("success")
			} else {
				logger.Warn().Int(log.Code, m.Code).Stringer(log.Duration, m.Duration).Msg("failed to handle")
			}
		})
	}
}

// Health it must be the last middleware.
// Needed to check for health from the discovery service.
func Health(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const health = `/health`
		if r.URL.Path != health {
			next.ServeHTTP(w, r)

			return
		}

		logger := zerolog.Ctx(r.Context())
		w.WriteHeader(http.StatusOK)
		logger.Info().Msg("handled discovery checker")
	})
}
