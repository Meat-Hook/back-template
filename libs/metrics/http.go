package metrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-openapi/loads"
	"github.com/prometheus/client_golang/prometheus"
)

// API contains main metrics for http methods.
type API struct {
	ReqInFlight prometheus.Gauge
	ReqTotal    *prometheus.CounterVec
	ReqDuration *prometheus.HistogramVec
}

// HTTP registers and returns common http metrics used by all
// services (namespace).
func HTTP(service string, swagger json.RawMessage) (metric API) {
	const subsystem = "api"

	metric.ReqInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: service,
			Subsystem: subsystem,
			Name:      "http_requests_in_flight",
			Help:      "Amount of currently processing API requests.",
		},
	)
	prometheus.DefaultRegisterer.MustRegister(metric.ReqInFlight)

	metric.ReqTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: service,
			Subsystem: subsystem,
			Name:      "http_requests_total",
			Help:      "Amount of processed API requests.",
		},
		[]string{MethodLabel, CodeLabel, ResourceLabel},
	)
	prometheus.DefaultRegisterer.MustRegister(metric.ReqTotal)

	metric.ReqDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: service,
			Subsystem: subsystem,
			Name:      "http_request_duration_seconds",
			Help:      "API request latency distributions.",
		},
		[]string{MethodLabel, CodeLabel, ResourceLabel},
	)
	prometheus.DefaultRegisterer.MustRegister(metric.ReqDuration)

	document, err := loads.Analyzed(swagger, "")
	if err != nil {
		panic(fmt.Errorf("analyzed swagger: %w", err))
	}

	// Initialized with codes returned by swagger and middleware
	// after metrics middleware (accessLog).
	codeLabels := [4]int{
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusUnprocessableEntity,
	}

	for method, resources := range document.Analyzer.Operations() {
		for resource, op := range resources {
			codes := append([]int{}, codeLabels[:]...)
			for code := range op.Responses.StatusCodeResponses {
				codes = append(codes, code)
			}
			for _, code := range codes {
				l := prometheus.Labels{
					ResourceLabel: resource,
					MethodLabel:   method,
					CodeLabel:     strconv.Itoa(code),
				}
				metric.ReqTotal.With(l)
				metric.ReqDuration.With(l)
			}
		}
	}

	return metric
}
