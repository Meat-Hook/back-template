package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/Meat-Hook/back-template/libs/log"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

const appVersion = `0.1.0`

type embeddedService interface {
	Name() string
	Flags() []cli.Flag
	RunServe(context.Context) error
}

var (
	output = zerolog.ConsoleWriter{
		Out:        os.Stdout,   // Standard output.
		NoColor:    false,       // Not useful.
		TimeFormat: time.RFC850, // Use for eyes.
	}
	embeddedServices = []embeddedService{}
)

func main() {
	logger := zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Caller().Timestamp().Logger()

	app := &cli.App{
		Name:                 filepath.Base(os.Args[0]),
		HelpName:             filepath.Base(os.Args[0]),
		Usage:                "Generate easy CRUD server.",
		Version:              appVersion,
		Commands:             []*cli.Command{},
		Flags:                []cli.Flag{},
		EnableBashCompletion: true,
		BashComplete:         cli.DefaultAppComplete,
		Action:               start,
		Reader:               os.Stdin,
		Writer:               os.Stdout,
		ErrWriter:            os.Stderr,
	}

	// service name -> duplicate
	duplicateService := make(map[string]bool)
	for _, service := range embeddedServices {
		name := service.Name()
		if duplicateService[name] {
			panic(fmt.Sprintf("duplicate service: %s", name))
		}
		duplicateService[name] = true

		app.Flags = append(app.Flags, service.Flags()...)
	}

	signals := make(chan os.Signal, 1)
	ctxParent := logger.WithContext(context.Background())
	ctx, cancel := signal.NotifyContext(ctxParent, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGTERM)
	go func() { <-signals; cancel() }()
	go forceShutdown(ctx)

	err := app.RunContext(ctx, os.Args)
	if err != nil {
		logger.Fatal().Err(err).Str(log.Version, appVersion).Msg("shutdown")
	}
}

func start(c *cli.Context) error {
	logger := zerolog.Ctx(c.Context)
	g, ctx := errgroup.WithContext(c.Context)

	for _, svc := range embeddedServices {
		svc := svc

		name := svc.Name()
		serviceLogger := logger.With().Str(log.Service, name).Logger()

		g.Go(func() error {
			return svc.RunServe(serviceLogger.WithContext(ctx))
		})
	}

	return g.Wait()
}

func forceShutdown(ctx context.Context) {
	const shutdownDelay = 15 * time.Second

	<-ctx.Done()
	time.Sleep(shutdownDelay)

	zerolog.Ctx(ctx).Fatal().Str(log.Version, appVersion).Msg("failed to graceful shutdown")
}
