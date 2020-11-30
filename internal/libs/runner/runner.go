// Package runner need for start server application.
package runner

import (
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	shutdownTimeout = time.Second * 15
)

// Standard ports.
const (
	WebServerPort    = 8080
	GRPCServerPort   = 8090
	MetricServerPort = 8100
)

// Start application services.
func Start(services ...func() error) error {
	group := errgroup.Group{}

	for i := range services {
		group.Go(services[i])
	}

	return group.Wait()
}
