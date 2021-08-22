package file

import (
	"context"
	"encoding/json"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"

	"github.com/Meat-Hook/back-template/cmd/file/internal/api/rpc"
	"github.com/Meat-Hook/back-template/cmd/file/internal/api/web"
	"github.com/Meat-Hook/back-template/cmd/file/internal/api/web/generated/restapi"
	"github.com/Meat-Hook/back-template/cmd/file/internal/app"
	"github.com/Meat-Hook/back-template/cmd/file/internal/services/repo"
	"github.com/Meat-Hook/back-template/libs/db"
	"github.com/Meat-Hook/back-template/libs/log"
	"github.com/Meat-Hook/back-template/libs/reflect"
	librpc "github.com/Meat-Hook/back-template/libs/rpc"
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
			GRPC   int `json:"grpc"`
		} `json:"port"`
	} `json:"server"`
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

	// Build contracts.
	r := repo.New(pg)

	module := app.New(r)

	grpcAPI := rpc.New(ctx, module, librpc.NewServerMetrics(reg, namespace))

	webMetric := libweb.NewMetric(reg, namespace, restapi.FlatSwaggerJSON)
	webAPI, err := web.New(ctx, module, &webMetric, web.Config{
		Host: s.cfg.Server.Host,
		Port: s.cfg.Server.Port.WEB,
	})
	if err != nil {
		return fmt.Errorf("web.New: %w", err)
	}

	return serve.Start(
		ctx,
		serve.Metrics(logger.With().Str(log.Subsystem, "metric").Logger(), s.cfg.Server.Host, s.cfg.Server.Port.Metric, reg),
		serve.HTTP(logger.With().Str(log.Subsystem, "web").Logger(), s.cfg.Server.Host, s.cfg.Server.Port.WEB, webAPI.GetHandler()),
		serve.GRPC(logger.With().Str(log.Subsystem, "grpc").Logger(), s.cfg.Server.Host, s.cfg.Server.Port.GRPC, grpcAPI),
	)
}
