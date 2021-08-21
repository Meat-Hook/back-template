package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"

	"github.com/Meat-Hook/back-template/cmd/file"
	"github.com/Meat-Hook/back-template/cmd/session"
	"github.com/Meat-Hook/back-template/cmd/user"
	"github.com/Meat-Hook/back-template/libs/log"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

const appVersion = `0.1.0`

type embeddedService interface {
	Name() string
	UnmarshalConfig(json.RawMessage) error
	RunServe(ctx context.Context, reg *prometheus.Registry, namespace string) error
}

//nolint:gochecknoglobals,nolintlint // By design.
var (
	embeddedServices = []embeddedService{
		&user.Service{},
		&session.Service{},
		&file.Service{},
	}
	cfgFilePath = &cli.StringFlag{
		Name:     "cfg",
		Aliases:  []string{"c"},
		Usage:    "config file path",
		EnvVars:  []string{"CONFIG_FILE_PATH"},
		Required: true,
	}
	// Service name -> json config.
	config map[string]json.RawMessage
)

func main() {
	logger := zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Caller().Timestamp().Logger()

	app := &cli.App{
		Name:                 filepath.Base(os.Args[0]),
		HelpName:             filepath.Base(os.Args[0]),
		Usage:                "Generate easy CRUD server.",
		Version:              appVersion,
		Commands:             []*cli.Command{},
		Flags:                []cli.Flag{cfgFilePath},
		EnableBashCompletion: true,
		BashComplete:         cli.DefaultAppComplete,
		Action:               start,
		Reader:               os.Stdin,
		Writer:               os.Stdout,
		ErrWriter:            os.Stderr,
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

	cfgFile, err := os.Open(c.String(cfgFilePath.Name))
	if err != nil {
		return fmt.Errorf("os.Open: %w", err)
	}

	err = json.NewDecoder(cfgFile).Decode(&config)
	if err != nil {
		return fmt.Errorf("json.NewDecoder.Decode: %w", err)
	}

	g, ctx := errgroup.WithContext(c.Context)

	for _, svc := range embeddedServices {
		svc := svc
		name := svc.Name()
		err := svc.UnmarshalConfig(config[name])
		if err != nil {
			return fmt.Errorf("svc.UnmarshalConfig: %w, json: %s, svc: %s", err, config[name], name)
		}

		g.Go(func() error {
			serviceLogger := logger.With().Str(log.Service, name).Logger()
			reg := prometheus.NewRegistry()

			err = svc.RunServe(serviceLogger.WithContext(ctx), reg, c.App.Name)
			if err != nil {
				return fmt.Errorf("svc.RunServe: %w", err)
			}

			return nil
		})
	}

	err = g.Wait()
	if err != nil {
		return fmt.Errorf("g.Wait: %w", err)
	}

	return nil
}

func forceShutdown(ctx context.Context) {
	const shutdownDelay = 15 * time.Second

	<-ctx.Done()
	time.Sleep(shutdownDelay)

	zerolog.Ctx(ctx).Fatal().Str(log.Version, appVersion).Msg("failed to graceful shutdown")
}
