// Package runner need for start server application.
package runner

import (
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	shutdownTimeout = time.Second * 15
)

// Start application services.
func Start(services ...func() error) error {
	group := errgroup.Group{}

	for i := range services {
		group.Go(services[i])
	}

	return group.Wait()
}
