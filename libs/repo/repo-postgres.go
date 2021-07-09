package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/rs/zerolog"

	"github.com/Meat-Hook/back-template/libs/log"
)

// Error names.
const (
	PostgresUniqueViolation     = "unique_violation"
	PostgresForeignKeyViolation = "foreign_key_violation"
)

// PostgresErrName returns true if err is PostgreSQL error with given name.
func PostgresErrName(err error, name string) bool {
	pqErr := new(pq.Error)
	return errors.As(err, &pqErr) && pqErr.Code.Name() == name
}

// PostgresConfig contains repo configuration.
type PostgresConfig struct {
	DSN        string
	MigrateDir string
	Metric     Metrics
}

// Postgres creates and returns new Repo.
// It will also run required DB migrations and connects to DB.
func Postgres(ctx context.Context, cfg PostgresConfig) (_ *Repo, err error) {
	logger := *zerolog.Ctx(ctx)

	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}
	defer func() {
		if err != nil {
			log.WarnIfFail(logger, db.Close)
		}
	}()

	err = db.PingContext(ctx)
	for err != nil {
		nextErr := db.PingContext(ctx)
		if errors.Is(nextErr, context.DeadlineExceeded) || errors.Is(nextErr, context.Canceled) {
			return nil, fmt.Errorf("db.Ping: %w", err)
		}
		err = nextErr
	}

	r := &Repo{
		db:     sqlx.NewDb(db, "postgres"),
		metric: cfg.Metric,
	}

	return r, nil
}
