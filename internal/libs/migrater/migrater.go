// Package migrater contains migrate module.
package migrater

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Meat-Hook/migrate/core"
	"github.com/Meat-Hook/migrate/filesystem"
	"github.com/Meat-Hook/migrate/repo"
	"github.com/rs/zerolog"
)

// Auto start automate migration to database.
func Auto(ctx context.Context, db *sql.DB, pathToMigrateDir string, logger zerolog.Logger) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	m := core.New(logger, filesystem.New(), repo.New(tx))
	err = m.Migrate(ctx, pathToMigrateDir, core.Config{Cmd: core.Up})
	if err != nil {
		return fmt.Errorf("migrate: %w, rollback: %s", err, tx.Rollback())
	}

	return tx.Commit()
}
