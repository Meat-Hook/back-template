package runner

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type Runner struct {
	runners []func() error
}

func New(runners ...func() error) *Runner {
	return &Runner{
		runners: runners,
	}
}

func (r *Runner) Run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	for i := range r.runners {
		group.Go(r.runners[i])
	}

	return group.Wait()
}
