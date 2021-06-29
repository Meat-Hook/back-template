package repo

import (
	"database/sql"
	"errors"

	app2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/app"
)

func convertErr(err error) error {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return app2.ErrNotFound
	default:
		return err
	}
}
