package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Meat-Hook/back-template/internal/libs/log"
	"github.com/Meat-Hook/back-template/internal/libs/metrics"
	"github.com/Meat-Hook/back-template/internal/libs/migrater"
	librpc "github.com/Meat-Hook/back-template/internal/libs/rpc"
	"github.com/Meat-Hook/back-template/internal/libs/runner"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/api/rpc"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/api/web"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/api/web/generated/restapi"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/app"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/auth"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/repo"
	wrapper "github.com/Meat-Hook/back-template/internal/modules/session/internal/users"
	users "github.com/Meat-Hook/back-template/internal/modules/user/client"
	"github.com/go-openapi/loads"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

var (
	logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	DBName = &cli.StringFlag{
		Name:     "db-name",
		Aliases:  []string{"n"},
		Usage:    "database name",
		EnvVars:  []string{"DB_NAME"},
		Required: true,
	}
	DBUser = &cli.StringFlag{
		Name:     "db-user",
		Aliases:  []string{"u"},
		Usage:    "database user",
		EnvVars:  []string{"DB_USER"},
		Required: true,
	}
	DBPass = &cli.StringFlag{
		Name:     "db-pass",
		Aliases:  []string{"p"},
		Usage:    "database password",
		EnvVars:  []string{"DB_PASS"},
		Required: true,
	}
	DBHost = &cli.StringFlag{
		Name:     "db-host",
		Aliases:  []string{"H"},
		Usage:    "database host",
		EnvVars:  []string{"DB_HOST"},
		Required: true,
	}
	DBPort = &cli.IntFlag{
		Name:     "db-port",
		Aliases:  []string{"P"},
		Usage:    "database port",
		EnvVars:  []string{"DB_PORT"},
		Required: true,
	}
	UserSrv = &cli.StringFlag{
		Name:     "user-srv",
		Usage:    "user server address",
		EnvVars:  []string{"USER_SRV"},
		Required: true,
	}
	Host = &cli.StringFlag{
		Name:    "hostname",
		Usage:   "service hostname",
		EnvVars: []string{"HOSTNAME"},
	}
	AuthKey = &cli.StringFlag{
		Name:     "auth-key",
		Usage:    "key for encrypt and decrypt tokens",
		EnvVars:  []string{"AUTH_KEY"},
		Required: true,
	}
	GRPCPort = &cli.IntFlag{
		Name:       "grpc-port",
		Usage:      "grpc service port",
		EnvVars:    []string{"GRPC_PORT"},
		Value:      runner.GRPCServerPort,
		Required:   true,
		HasBeenSet: true,
	}
	HTTPPort = &cli.IntFlag{
		Name:       "http-port",
		Usage:      "http service port",
		EnvVars:    []string{"HTTP_PORT"},
		Value:      runner.WebServerPort,
		Required:   true,
		HasBeenSet: true,
	}
	MetricPort = &cli.IntFlag{
		Name:       "metric-port",
		Usage:      "metric service port",
		EnvVars:    []string{"METRIC_PORT"},
		Value:      runner.MetricServerPort,
		Required:   true,
		HasBeenSet: true,
	}
	Migrate = &cli.BoolFlag{
		Name:       "migrate",
		Usage:      "start automatic migrate to database",
		EnvVars:    []string{"MIGRATE"},
		Value:      false,
		Required:   true,
		HasBeenSet: true,
	}
	MigrateDir = &cli.StringFlag{
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
			DBName, DBPass, DBUser, DBPort, DBHost, AuthKey,
			UserSrv, Host, GRPCPort, HTTPPort, MetricPort,
			Migrate, MigrateDir,
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
	host, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("os hostname: %w", err)
	}

	if val := c.String(Host.Name); val != "" {
		host = val
	}

	// init database connection
	dbMetric := metrics.DB(name, metrics.MethodsOf(&repo.Repo{})...)
	db, err := sqlx.Connect(dbDriver, fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable", c.String(DBHost.Name), c.Int(DBPort.Name), c.String(DBUser.Name),
		c.String(DBPass.Name), c.String(DBName.Name)))
	if err != nil {
		return fmt.Errorf("DB connect: %w", err)
	}
	defer log.WarnIfFail(logger, db.Close)

	if c.Bool(Migrate.Name) {
		err := migrater.Auto(c.Context, db.DB, c.String(MigrateDir.Name), logger)
		if err != nil {
			return fmt.Errorf("start auto migration: %w", err)
		}
	}

	grpcConn, err := librpc.Client(c.Context, c.String(UserSrv.Name))
	if err != nil {
		return fmt.Errorf("build lib rpc: %w", err)
	}
	usersSvcClient := users.New(grpcConn)

	r := repo.New(db, &dbMetric)
	auth := auth.New(c.String(AuthKey.Name))

	module := app.New(r, wrapper.New(usersSvcClient), auth, idGen{})

	apiMetric := metrics.HTTP(name, restapi.FlatSwaggerJSON)
	internalAPI := rpc.New(module, librpc.Server(logger))
	externalAPI, err := web.New(module, logger, &apiMetric, web.Config{
		Host: host,
		Port: c.Int(HTTPPort.Name),
	})
	if err != nil {
		return fmt.Errorf("build external api: %w", err)
	}

	return runner.Start(
		c.Context,
		runner.GRPC(logger.With().Str(log.Name, "GRPC").Logger(), internalAPI, host, c.Int(GRPCPort.Name)),
		runner.HTTP(logger.With().Str(log.Name, "HTTP").Logger(), externalAPI, host, c.Int(HTTPPort.Name)),
		runner.Metric(logger.With().Str(log.Name, "Metric").Logger(), host, c.Int(MetricPort.Name)),
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

var _ app.ID = &idGen{}

type idGen struct{}

// New for implements app.ID.
func (i idGen) New() string {
	return xid.New().String()
}
