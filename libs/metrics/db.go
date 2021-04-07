package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Database contains main metrics for db methods.
type Database struct {
	callTotal    *prometheus.CounterVec
	callErrTotal *prometheus.CounterVec
	callDuration *prometheus.HistogramVec
}

// DB registers and returns common repo metrics used by all services (namespace).
func DB(service string, registerMethods ...string) (metric Database) {
	const subsystem = "repo"

	metric.callTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: service,
			Subsystem: subsystem,
			Name:      "call_total",
			Help:      "Amount of repo calls.",
		},
		[]string{MethodLabel},
	)
	prometheus.DefaultRegisterer.MustRegister(metric.callTotal)
	metric.callErrTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: service,
			Subsystem: subsystem,
			Name:      "errors_total",
			Help:      "Amount of repo errors.",
		},
		[]string{MethodLabel},
	)
	prometheus.DefaultRegisterer.MustRegister(metric.callErrTotal)
	metric.callDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: service,
			Subsystem: subsystem,
			Name:      "call_duration_seconds",
			Help:      "Repo call latency.",
		},
		[]string{MethodLabel},
	)
	prometheus.DefaultRegisterer.MustRegister(metric.callDuration)

	for _, methodName := range registerMethods {
		l := prometheus.Labels{
			MethodLabel: methodName,
		}
		metric.callTotal.With(l)
		metric.callErrTotal.With(l)
		metric.callDuration.With(l)
	}

	return metric
}

// Collect call callback and collects launch metrics.
func (m Database) Collect(f func() error) error {
	start := time.Now()
	method := CallerMethodName()
	l := prometheus.Labels{MethodLabel: method}

	m.callTotal.With(l).Inc()
	m.callDuration.With(l).Observe(time.Since(start).Seconds())

	err := f()
	if err != nil {
		m.callErrTotal.With(l).Inc()
	} else if err := recover(); err != nil {
		m.callErrTotal.With(l).Inc()
		panic(err)
	}

	return err
}
