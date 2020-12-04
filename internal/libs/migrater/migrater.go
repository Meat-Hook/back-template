package migrater

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Meat-Hook/migrate/core"
	"github.com/Meat-Hook/migrate/fs"
	"github.com/Meat-Hook/migrate/migrater"
	"github.com/rs/zerolog"
)

type wrapperForLog struct {
	logger zerolog.Logger
}

func (w wrapperForLog) Info(i ...interface{}) {
	w.logger.Info().Msgf("%s", i)
}

// Auto start automate migration to database.
func Auto(ctx context.Context, db *sql.DB, pathToMigrateDir string, logger zerolog.Logger) error {
	m := core.New(fs.New(), migrater.New(db, wrapperForLog{logger: logger}))

	err := m.Migrate(ctx, pathToMigrateDir, core.Config{Cmd: core.Up})
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	return nil
}
