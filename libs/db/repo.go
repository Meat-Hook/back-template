// Package db provide helpers for Data Access Layer.
package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"github.com/Meat-Hook/back-template/libs/log"
	"github.com/Meat-Hook/back-template/libs/reflect"
)

// DB provides access to storage.
type DB struct {
	conn     *sqlx.DB
	metric Metrics
}

// NoTx provides DAL method wrapper with:
// - general metrics for DAL methods,
// - wrapping errors with DAL method name.
func (d *DB) NoTx(f func(db *sqlx.DB) error) (err error) {
	methodName := reflect.CallerMethodName(1)
	return d.metric.instrument(methodName, func() error {
		err := f(d.conn)
		if err != nil {
			err = fmt.Errorf("%s: %w", methodName, err)
		}
		return err
	})()
}

// Tx provides DAL method wrapper with:
// - general metrics for DAL methods,
// - wrapping errors with DAL method name,
// - transaction.
func (d *DB) Tx(ctx context.Context, opts *sql.TxOptions, f func(*sqlx.Tx) error) (err error) {
	methodName := reflect.CallerMethodName(1)
	return d.metric.instrument(methodName, func() error {
		tx, err := d.conn.BeginTxx(ctx, opts)
		if err == nil { //nolint:nestif // No idea how to simplify.
			defer func() {
				if err := recover(); err != nil {
					if err := tx.Rollback(); err != nil {
						logger := zerolog.Ctx(ctx)
						logger.Warn().Err(err).Str(log.DBMethod, methodName).Msg("failed to tx.Rollback")
					}
					panic(err)
				}
			}()
			err = f(tx)
			if err == nil {
				err = tx.Commit()
			} else if err := tx.Rollback(); err != nil {
				logger := zerolog.Ctx(ctx)
				logger.Warn().Err(err).Str(log.DBMethod, methodName).Msg("failed to tx.Rollback")
			}
		}
		if err != nil {
			err = fmt.Errorf("%s: %w", methodName, err)
		}
		return err
	})()
}
