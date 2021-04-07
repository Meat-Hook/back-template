// Package runner need for start server application.
package runner

import (
	"context"
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
func Start(ctx context.Context, services ...func(context.Context) error) error {
	group, groupCtx := errgroup.WithContext(ctx)

	for i := range services {
		i := i
		group.Go(func() error {
			return services[i](groupCtx)
		})
	}

	return group.Wait()
}
