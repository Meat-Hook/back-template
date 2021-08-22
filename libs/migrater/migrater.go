// Package migrater contains migrate module.
package migrater

import (
	"context"
	"database/sql"
	"fmt"

	migrate "github.com/Meat-Hook/migrate/core"
	"github.com/Meat-Hook/migrate/filesystem"
	"github.com/Meat-Hook/migrate/repo"
	"github.com/rs/zerolog"
)

// Auto start automate migration to database.
func Auto(ctx context.Context, db *sql.DB, dir string) error {
	logger := zerolog.Ctx(ctx)
	logger.Info().Msg("started migration...")
	defer logger.Info().Msg("finished migration")

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	m := migrate.New(*logger, filesystem.New(), repo.New(tx))
	err = m.Migrate(ctx, dir, migrate.Config{Cmd: migrate.Up})
	if err != nil {
		return fmt.Errorf("m.Migrate: %w, tx.Rollback: %s", err, tx.Rollback())
	}

	return tx.Commit()
}
