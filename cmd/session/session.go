package session

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"

	"github.com/Meat-Hook/back-template/cmd/session/internal/api/rpc"
	"github.com/Meat-Hook/back-template/cmd/session/internal/app"
	"github.com/Meat-Hook/back-template/cmd/session/internal/auth"
	"github.com/Meat-Hook/back-template/cmd/session/internal/services/repo"
	"github.com/Meat-Hook/back-template/libs/db"
	"github.com/Meat-Hook/back-template/libs/log"
	"github.com/Meat-Hook/back-template/libs/reflect"
	"github.com/Meat-Hook/back-template/libs/serve"
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
			GRPC   int `json:"grpc"`
			Metric int `json:"metric"`
		} `json:"port"`
	} `json:"server"`
	AuthKey string `json:"auth_key"`
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
	return json.Unmarshal(buf, &s.cfg)
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
	authModule := auth.New(s.cfg.AuthKey)

	module := app.New(r, authModule, idGenerator{})

	grpcAPI := rpc.New(ctx, reg, namespace, module)

	return serve.Start(
		ctx,
		serve.Metrics(logger.With().Str(log.Subsystem, "metric").Logger(), s.cfg.Server.Host, s.cfg.Server.Port.Metric, reg),
		serve.GRPC(logger.With().Str(log.Subsystem, "grpc").Logger(), s.cfg.Server.Host, s.cfg.Server.Port.GRPC, grpcAPI),
	)
}

var _ app.ID = &idGenerator{}

type idGenerator struct{}

// New implements app.ID.
func (idGenerator) New() uuid.UUID {
	return uuid.Must(uuid.NewV4())
}
