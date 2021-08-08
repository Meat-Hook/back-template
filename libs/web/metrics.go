package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-openapi/loads"
	"github.com/prometheus/client_golang/prometheus"
)

// Metric contains main metrics for web methods.
type Metric struct {
	ReqInFlight prometheus.Gauge
	ReqTotal    *prometheus.CounterVec
	ReqDuration *prometheus.HistogramVec
}

// NewMetric registers and returns common web metrics used by all
// services (namespace).
func NewMetric(reg *prometheus.Registry, namespace string, swagger json.RawMessage) (metric Metric) {
	const subsystem = "web"

	// Labels.
	const (
		resourceLabel = "resource"
		methodLabel   = "method"
		codeLabel     = "code"
	)

	metric.ReqInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "http_requests_in_flight",
			Help:      "Amount of currently processing API requests.",
		},
	)
	reg.MustRegister(metric.ReqInFlight)

	metric.ReqTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "http_requests_total",
			Help:      "Amount of processed API requests.",
		},
		[]string{methodLabel, codeLabel, resourceLabel},
	)
	reg.MustRegister(metric.ReqTotal)

	metric.ReqDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "http_request_duration_seconds",
			Help:      "API request latency distributions.",
		},
		[]string{methodLabel, codeLabel, resourceLabel},
	)
	reg.MustRegister(metric.ReqDuration)

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
					resourceLabel: resource,
					methodLabel:   method,
					codeLabel:     strconv.Itoa(code),
				}
				metric.ReqTotal.With(l)
				metric.ReqDuration.With(l)
			}
		}
	}

	return metric
}
