package user

import (
	"context"
	"encoding/json"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"

	file_client "github.com/Meat-Hook/back-template/cmd/file/client"
	session_client "github.com/Meat-Hook/back-template/cmd/session/client"
	"github.com/Meat-Hook/back-template/cmd/user/internal/api/web"
	"github.com/Meat-Hook/back-template/cmd/user/internal/api/web/generated/restapi"
	"github.com/Meat-Hook/back-template/cmd/user/internal/app"
	"github.com/Meat-Hook/back-template/cmd/user/internal/services/file"
	"github.com/Meat-Hook/back-template/cmd/user/internal/services/repo"
	"github.com/Meat-Hook/back-template/cmd/user/internal/services/session"
	"github.com/Meat-Hook/back-template/libs/db"
	"github.com/Meat-Hook/back-template/libs/hash"
	"github.com/Meat-Hook/back-template/libs/log"
	"github.com/Meat-Hook/back-template/libs/reflect"
	"github.com/Meat-Hook/back-template/libs/rpc"
	"github.com/Meat-Hook/back-template/libs/serve"
	libweb "github.com/Meat-Hook/back-template/libs/web"
)

type config struct {
	DB struct {
		DSN        string `json:"dsn"`
		Driver     string `json:"driver"`
		MigrateDir string `json:"migrate_dir"`
	} `json:"db"`
	Server struct {
		Host string `json:"host"`
		Port struct {
			WEB    int `json:"web"`
			Metric int `json:"metric"`
		} `json:"port"`
	} `json:"server"`
	Services struct {
		SessionAddr string `json:"session_addr"`
		FileAddr    string `json:"file_addr"`
	} `json:"services"`
}

const version = "v0.1.0"

// Service module implementation.
type Service struct {
	cfg config
}

// Name implements main.embeddedService.
func (s *Service) Name() string {
	return reflect.CallerPkg(0)
}

// UnmarshalConfig implements main.embeddedService.
func (s *Service) UnmarshalConfig(buf json.RawMessage) error {
	err := json.Unmarshal(buf, &s.cfg)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	return nil
}

// RunServe implements main.embeddedService.
func (s *Service) RunServe(ctx context.Context, reg *prometheus.Registry, namespace string) error {
	logger := zerolog.Ctx(ctx).With().Str(log.Version, version).Logger()

	dbMetric := db.NewMetrics(reg, namespace, &repo.Repo{})
	pg, err := db.Postgres(logger.WithContext(ctx), db.PostgresConfig{
		DSN:        s.cfg.DB.DSN,
		MigrateDir: s.cfg.DB.MigrateDir,
		Metric:     *dbMetric,
	})
	if err != nil {
		return fmt.Errorf("db.Postgres: %w", err)
	}

	grpcClientMetric := rpc.NewClientMetrics(reg, namespace)
	grpcConnSession, err := rpc.Dial(ctx, logger, s.cfg.Services.SessionAddr, grpcClientMetric)
	if err != nil {
		return fmt.Errorf("rpc.Dial: %w", err)
	}

	grpcConnFile, err := rpc.Dial(ctx, logger, s.cfg.Services.FileAddr, grpcClientMetric)
	if err != nil {
		return fmt.Errorf("rpc.Dial: %w", err)
	}

	// Build contracts.
	sessionSvcClient := session.New(session_client.New(grpcConnSession))
	fileSvcClient := file.New(file_client.New(grpcConnFile))
	r := repo.New(pg)
	hasher := hash.New()

	module := app.New(r, hasher, sessionSvcClient, fileSvcClient)

	webMetric := libweb.NewMetric(reg, namespace, restapi.FlatSwaggerJSON)
	webAPI, err := web.New(ctx, module, &webMetric, web.Config{
		Host: s.cfg.Server.Host,
		Port: s.cfg.Server.Port.WEB,
	})
	if err != nil {
		return fmt.Errorf("web.New: %w", err)
	}

	err = serve.Start(
		ctx,
		serve.Metrics(logger.With().Str(log.Subsystem, "metric").Logger(), s.cfg.Server.Host, s.cfg.Server.Port.Metric, reg),
		serve.HTTP(logger.With().Str(log.Subsystem, "web").Logger(), s.cfg.Server.Host, s.cfg.Server.Port.WEB, webAPI.GetHandler()),
	)
	if err != nil {
		return fmt.Errorf("serve.Start: %w", err)
	}

	return nil
}
