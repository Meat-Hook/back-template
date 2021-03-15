package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Meat-Hook/back-template/internal/libs/hash"
	"github.com/Meat-Hook/back-template/internal/libs/log"
	"github.com/Meat-Hook/back-template/internal/libs/metrics"
	"github.com/Meat-Hook/back-template/internal/libs/migrater"
	librpc "github.com/Meat-Hook/back-template/internal/libs/rpc"
	"github.com/Meat-Hook/back-template/internal/libs/runner"
	session "github.com/Meat-Hook/back-template/microservices/session/client"
	"github.com/Meat-Hook/back-template/microservices/user/internal/api/rpc"
	"github.com/Meat-Hook/back-template/microservices/user/internal/api/web"
	"github.com/Meat-Hook/back-template/microservices/user/internal/api/web/generated/restapi"
	"github.com/Meat-Hook/back-template/microservices/user/internal/app"
	"github.com/Meat-Hook/back-template/microservices/user/internal/notification"
	"github.com/Meat-Hook/back-template/microservices/user/internal/repo"
	wrapper "github.com/Meat-Hook/back-template/microservices/user/internal/session"
	"github.com/go-openapi/loads"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

var (
	logger = zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Caller().Timestamp().Logger()

	dbName = &cli.StringFlag{
		Name:     "db-name",
		Aliases:  []string{"n"},
		Usage:    "database name",
		EnvVars:  []string{"DB_NAME"},
		Required: true,
	}
	dbUser = &cli.StringFlag{
		Name:     "db-user",
		Aliases:  []string{"u"},
		Usage:    "database user",
		EnvVars:  []string{"DB_USER"},
		Required: true,
	}
	dbPass = &cli.StringFlag{
		Name:     "db-pass",
		Aliases:  []string{"p"},
		Usage:    "database password",
		EnvVars:  []string{"DB_PASS"},
		Required: true,
	}
	dbHost = &cli.StringFlag{
		Name:     "db-host",
		Aliases:  []string{"H"},
		Usage:    "database host",
		EnvVars:  []string{"DB_HOST"},
		Required: true,
	}
	dbPort = &cli.IntFlag{
		Name:     "db-port",
		Aliases:  []string{"P"},
		Usage:    "database port",
		EnvVars:  []string{"DB_PORT"},
		Required: true,
	}
	natsAddr = &cli.StringFlag{
		Name:     "nats",
		Usage:    "nats server address",
		EnvVars:  []string{"NATS"},
		Required: true,
	}
	sessionSrv = &cli.StringFlag{
		Name:     "session-srv",
		Usage:    "session server address",
		EnvVars:  []string{"SESSION_SRV"},
		Required: true,
	}
	host = &cli.StringFlag{
		Name:    "hostname",
		Usage:   "service hostname",
		EnvVars: []string{"HOSTNAME"},
	}
	grpcPort = &cli.IntFlag{
		Name:       "grpc-port",
		Usage:      "grpc service port",
		EnvVars:    []string{"GRPC_PORT"},
		Value:      runner.GRPCServerPort,
		Required:   true,
		HasBeenSet: true,
	}
	httpPort = &cli.IntFlag{
		Name:       "http-port",
		Usage:      "http service port",
		EnvVars:    []string{"HTTP_PORT"},
		Value:      runner.WebServerPort,
		Required:   true,
		HasBeenSet: true,
	}
	metricPort = &cli.IntFlag{
		Name:       "metric-port",
		Usage:      "metric service port",
		EnvVars:    []string{"METRIC_PORT"},
		Value:      runner.MetricServerPort,
		Required:   true,
		HasBeenSet: true,
	}
	migrate = &cli.BoolFlag{
		Name:       "migrate",
		Usage:      "start automatic migrate to database",
		EnvVars:    []string{"MIGRATE"},
		Value:      false,
		Required:   true,
		HasBeenSet: true,
	}
	migrateDir = &cli.StringFlag{
		Name:       "migrate-dir",
		Usage:      "path to database migration",
		EnvVars:    []string{"MIGRATE_DIR"},
		Value:      "migrate/",
		Required:   true,
		HasBeenSet: true,
	}

	author1 = &cli.Author{
		Name:  "Edgar Sipki",
		Email: "edo7796@yahoo.com",
	}

	team = []*cli.Author{author1}

	version = &cli.Command{
		Name:         "version",
		Aliases:      []string{"v"},
		Usage:        "Get service version.",
		Description:  "Command for getting service version.",
		BashComplete: cli.DefaultAppComplete,
		Action: func(context *cli.Context) error {
			doc, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
			if err != nil {
				logger.Fatal().Err(err).Msg("failed to get app version")
			}

			logger.Info().Str("version", doc.Version()).Send()

			return nil
		},
	}
)

func main() {
	doc, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to get app version")
	}

	application := &cli.App{
		Name:        filepath.Base(os.Args[0]),
		HelpName:    filepath.Base(os.Args[0]),
		Usage:       "Microservice for working with user info.",
		Description: "Microservice for working with user info.",
		Commands:    []*cli.Command{version},
		Flags: []cli.Flag{
			dbName, dbPass, dbUser, dbPort, dbHost, natsAddr,
			sessionSrv, host, grpcPort, httpPort, metricPort,
			migrate, migrateDir,
		},
		Version:              doc.Spec().Info.Version,
		EnableBashCompletion: true,
		BashComplete:         cli.DefaultAppComplete,
		Action:               start,
		Authors:              team,
		Writer:               os.Stdout,
		ErrWriter:            os.Stderr,
	}

	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGTERM)
	go func() { <-signals; cancel() }()
	go forceShutdown(ctx)

	err = application.RunContext(logger.WithContext(ctx), os.Args)
	if err != nil {
		logger.Fatal().Err(err).Msg("service shutdown")
	}
}

const (
	name     = `user` // Service name.
	dbDriver = `postgres`
)

func start(c *cli.Context) error {
	appHost, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("os hostname: %w", err)
	}

	if val := c.String(host.Name); val != "" {
		appHost = val
	}

	// init database connection
	dbMetric := metrics.DB(name, metrics.MethodsOf(&repo.Repo{})...)
	db, err := sqlx.Connect(dbDriver, fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable", c.String(dbHost.Name), c.Int(dbPort.Name), c.String(dbUser.Name),
		c.String(dbPass.Name), c.String(dbName.Name)))
	if err != nil {
		return fmt.Errorf("DB connect: %w", err)
	}
	defer log.WarnIfFail(logger, db.Close)

	if c.Bool(migrate.Name) {
		err := migrater.Auto(c.Context, logger.With().Str(log.Name, "migrate").Logger(), db.DB, c.String(migrateDir.Name))
		if err != nil {
			return fmt.Errorf("start auto migration: %w", err)
		}
	}

	natsConn, err := nats.Connect(c.String(natsAddr.Name))
	if err != nil {
		return fmt.Errorf("nats connect: %w", err)
	}
	defer log.WarnIfFail(logger, natsConn.Drain)
	defer natsConn.Close()

	grpcConn, err := librpc.Client(c.Context, c.String(sessionSrv.Name))
	if err != nil {
		return fmt.Errorf("build lib rpc: %w", err)
	}
	sessionSvcClient := session.New(grpcConn)

	r := repo.New(db, &dbMetric)
	hasher := hash.New()
	queueNotification := notification.New(natsConn)

	module := app.New(r, hasher, queueNotification, wrapper.New(sessionSvcClient))

	apiMetric := metrics.HTTP(name, restapi.FlatSwaggerJSON)
	internalAPI := rpc.New(module, librpc.Server(logger))
	externalAPI, err := web.New(module, logger, &apiMetric, web.Config{
		Host: appHost,
		Port: c.Int(httpPort.Name),
	})
	if err != nil {
		return fmt.Errorf("build external api: %w", err)
	}

	return runner.Start(
		c.Context,
		runner.GRPC(logger.With().Str(log.Name, "GRPC").Logger(), internalAPI, appHost, c.Int(grpcPort.Name)),
		runner.HTTP(logger.With().Str(log.Name, "HTTP").Logger(), externalAPI, appHost, c.Int(httpPort.Name)),
		runner.Metric(logger.With().Str(log.Name, "Metric").Logger(), appHost, c.Int(metricPort.Name)),
	)
}

func forceShutdown(ctx context.Context) {
	const shutdownDelay = 15 * time.Second

	<-ctx.Done()
	time.Sleep(shutdownDelay)
	doc, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to get app version")
	}

	logger.Fatal().Str("version", doc.Version()).Msg("failed to graceful shutdown")
}
