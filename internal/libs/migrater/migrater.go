// Package migrater contains migrate module.
package migrater

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	migrate "github.com/Meat-Hook/migrate/core"
	"github.com/Meat-Hook/migrate/filesystem"
	"github.com/Meat-Hook/migrate/repo"
	"github.com/rs/zerolog"
)

// Auto start automate migration to database.
func Auto(ctx context.Context, logger zerolog.Logger, db *sql.DB, dir string, fs fs.FS) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	m := migrate.New(logger, filesystem.New(), repo.New(tx))
	err = m.Migrate(ctx, dir, migrate.Config{Cmd: migrate.Up}, migrate.WithCustomFS(fs))
	if err != nil {
		return fmt.Errorf("migrate: %w, rollback: %s", err, tx.Rollback())
	}

	return tx.Commit()
}
