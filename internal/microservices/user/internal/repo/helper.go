package repo

import (
	"database/sql"
	"errors"

	"github.com/Meat-Hook/back-template/internal/microservices/user/internal/app"
)

func convertErr(err error) error {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return app.ErrNotFound
	default:
		return err
	}
}
