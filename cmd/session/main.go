package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	rpc3 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/api/rpc"
	web2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/api/web"
	restapi2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/api/web/generated/restapi"
	app2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/app"
	auth2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/auth"
	repo2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/repo"
	users2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/users"
	"github.com/Meat-Hook/back-template/internal/cmd/user/client"
	log2 "github.com/Meat-Hook/back-template/internal/libs/log"
	metrics2 "github.com/Meat-Hook/back-template/internal/libs/metrics"
	migrater2 "github.com/Meat-Hook/back-template/internal/libs/migrater"
	rpc2 "github.com/Meat-Hook/back-template/internal/libs/rpc"
	runner2 "github.com/Meat-Hook/back-template/internal/libs/runner"
	"github.com/go-openapi/loads"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/xid"
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
	dbSSLMode = &cli.StringFlag{
		Name:     "db-ssl",
		Aliases:  []string{"S"},
		Usage:    "database ssl mode",
		EnvVars:  []string{"DB_SSL_MODE"},
		Required: true,
	}
	dbPort = &cli.IntFlag{
		Name:     "db-port",
		Aliases:  []string{"P"},
		Usage:    "database port",
		EnvVars:  []string{"DB_PORT"},
		Required: true,
	}
	userSrv = &cli.StringFlag{
		Name:     "user-srv",
		Usage:    "user server address",
		EnvVars:  []string{"USER_SRV"},
		Required: true,
	}
	host = &cli.StringFlag{
		Name:    "hostname",
		Usage:   "service hostname",
		EnvVars: []string{"HOSTNAME"},
	}
	authKey = &cli.StringFlag{
		Name:     "auth-key",
		Usage:    "key for encrypt and decrypt tokens",
		EnvVars:  []string{"AUTH_KEY"},
		Required: true,
	}
	grpcPort = &cli.IntFlag{
		Name:       "grpc-port",
		Usage:      "grpc service port",
		EnvVars:    []string{"GRPC_PORT"},
		Value:      runner2.GRPCServerPort,
		Required:   true,
		HasBeenSet: true,
	}
	httpPort = &cli.IntFlag{
		Name:       "http-port",
		Usage:      "http service port",
		EnvVars:    []string{"HTTP_PORT"},
		Value:      runner2.WebServerPort,
		Required:   true,
		HasBeenSet: true,
	}
	metricPort = &cli.IntFlag{
		Name:       "metric-port",
		Usage:      "metric service port",
		EnvVars:    []string{"METRIC_PORT"},
		Value:      runner2.MetricServerPort,
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
			doc, err := loads.Analyzed(restapi2.FlatSwaggerJSON, "2.0")
			if err != nil {
				logger.Fatal().Err(err).Msg("failed to get app version")
			}

			logger.Info().Str("version", doc.Version()).Send()

			return nil
		},
	}
)

func main() {
	doc, err := loads.Analyzed(restapi2.FlatSwaggerJSON, "2.0")
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
			dbName, dbPass, dbUser, dbPort, dbHost, authKey, dbSSLMode,
			userSrv, host, grpcPort, httpPort, metricPort,
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
	name     = `session`
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
	dbMetric := metrics2.DB(name, metrics2.MethodsOf(&repo2.Repo{})...)
	db, err := sqlx.Connect(dbDriver, fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s", c.String(dbHost.Name), c.Int(dbPort.Name), c.String(dbUser.Name),
		c.String(dbPass.Name), c.String(dbName.Name), c.String(dbSSLMode.Name)))
	if err != nil {
		return fmt.Errorf("DB connect: %w", err)
	}
	defer log2.WarnIfFail(logger, db.Close)

	if c.Bool(migrate.Name) {
		err := migrater2.Auto(c.Context, logger.With().Str(log2.Name, "migrate").Logger(), db.DB, c.String(migrateDir.Name))
		if err != nil {
			return fmt.Errorf("start auto migration: %w", err)
		}
	}

	grpcConn, err := rpc2.Client(c.Context, c.String(userSrv.Name))
	if err != nil {
		return fmt.Errorf("build lib rpc: %w", err)
	}
	usersSvcClient := client.New(grpcConn)

	r := repo2.New(db, &dbMetric)
	authModule := auth2.New(c.String(authKey.Name))

	module := app2.New(r, users2.New(usersSvcClient), authModule, idGen{})

	apiMetric := metrics2.HTTP(name, restapi2.FlatSwaggerJSON)
	internalAPI := rpc3.New(module, rpc2.Server(logger))
	externalAPI, err := web2.New(module, logger, &apiMetric, web2.Config{
		Host: appHost,
		Port: c.Int(httpPort.Name),
	})
	if err != nil {
		return fmt.Errorf("build external api: %w", err)
	}

	return runner2.Start(
		c.Context,
		runner2.GRPC(logger.With().Str(log2.Name, "GRPC").Logger(), internalAPI, appHost, c.Int(grpcPort.Name)),
		runner2.HTTP(logger.With().Str(log2.Name, "HTTP").Logger(), externalAPI, appHost, c.Int(httpPort.Name)),
		runner2.Metric(logger.With().Str(log2.Name, "Metric").Logger(), appHost, c.Int(metricPort.Name)),
	)
}

func forceShutdown(ctx context.Context) {
	const shutdownDelay = 15 * time.Second

	<-ctx.Done()
	time.Sleep(shutdownDelay)
	doc, err := loads.Analyzed(restapi2.FlatSwaggerJSON, "2.0")
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to get app version")
	}

	logger.Fatal().Str("version", doc.Version()).Msg("failed to graceful shutdown")
}

var _ app2.ID = &idGen{}

type idGen struct{}

// New for implements app.ID.
func (i idGen) New() string {
	return xid.New().String()
}
