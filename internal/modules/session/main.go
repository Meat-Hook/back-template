package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/Meat-Hook/back-template/internal/libs/discovery"
	"github.com/Meat-Hook/back-template/internal/libs/log"
	"github.com/Meat-Hook/back-template/internal/libs/metrics"
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
	logger = zerolog.New(os.Stdout)

	discoveryFlg = &cli.StringFlag{
		Name:     "discovery",
		Usage:    "service discovery address for get config",
		EnvVars:  []string{"DISCOVERY"},
		Required: true,
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
		Name:                 filepath.Base(os.Args[0]),
		HelpName:             filepath.Base(os.Args[0]),
		Usage:                "Microservice for working with user info.",
		Description:          "Microservice for working with user info.",
		Commands:             []*cli.Command{version},
		Flags:                []cli.Flag{discoveryFlg},
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

type config struct {
	DB struct {
		Name     string `json:"name"`
		User     string `json:"user"`
		Password string `json:"password"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
	} `json:"db"`
	HTTP struct {
		Port int `json:"port"`
	} `json:"http"`
	GRPC struct {
		Port int `json:"port"`
	} `json:"grpc"`
	Auth struct {
		Key string `json:"key"`
	} `json:"auth"`
	Metric struct {
		Port int `json:"port"`
	} `json:"metric"`
	UserSvc struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"user_svc"`
}

const (
	name     = `session`
	dbDriver = `postgres`
)

func start(c *cli.Context) error {
	// init config
	discoveryClient, err := discovery.New(c.String(discoveryFlg.Name))
	if err != nil {
		return fmt.Errorf("init discovery: %w", err)
	}

	// build config
	cfg := config{}
	err = discoveryClient.Config(c.Context, name, &cfg)
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	host, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("os hostname: %w", err)
	}

	serverIP, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		return fmt.Errorf("resolve ip addr: %w", err)
	}

	doc, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to parse app doc")
	}

	serviceID := name + ":" + doc.Spec().Info.Version
	err = discoveryClient.Register(name, serviceID, serverIP.IP, cfg.HTTP.Port, doc.Spec().Info.Version)
	if err != nil {
		return fmt.Errorf("register service: %w", err)
	}
	defer func() {
		err := discoveryClient.Deregister(serviceID)
		if err != nil {
			logger.Error().Err(err).Str("id", serviceID).Msg("deregister service")
		}
	}()

	// init database connection
	dbMetric := metrics.DB(name, metrics.MethodsOf(&repo.Repo{})...)
	db, err := sqlx.Connect(dbDriver, fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable", cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name))
	if err != nil {
		return fmt.Errorf("DB connect: %w", err)
	}
	defer log.WarnIfFail(logger, db.Close)

	grpcConn, err := librpc.Client(c.Context, net.JoinHostPort(cfg.UserSvc.Host, strconv.Itoa(cfg.UserSvc.Port)))
	if err != nil {
		return fmt.Errorf("build lib rpc: %w", err)
	}
	usersSvcClient := users.New(grpcConn)

	r := repo.New(db, &dbMetric)
	auth := auth.New(cfg.Auth.Key)

	module := app.New(r, wrapper.New(usersSvcClient), auth, idGen{})

	apiMetric := metrics.HTTP(name, restapi.FlatSwaggerJSON)
	internalAPI := rpc.New(module, librpc.Server(logger))
	externalAPI, err := web.New(module, logger, &apiMetric, web.Config{
		Host: serverIP.IP.String(),
		Port: cfg.HTTP.Port,
	})
	if err != nil {
		return fmt.Errorf("build external api: %w", err)
	}

	return runner.Start(
		runner.GRPC(c.Context, logger.With().Str(log.Name, "GRPC").Logger(), internalAPI, serverIP.IP, cfg.GRPC.Port),
		runner.HTTP(c.Context, logger.With().Str(log.Name, "HTTP").Logger(), externalAPI, serverIP.IP, cfg.HTTP.Port),
		runner.Metric(c.Context, logger.With().Str(log.Name, "Metric").Logger(), serverIP.IP, cfg.Metric.Port),
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
