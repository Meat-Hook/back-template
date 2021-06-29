package repo

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	app2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/app"
	"github.com/lib/pq"
)

const (
	duplEmail    = "users_email_key"
	duplUsername = "users_name_key"
)

func convertErr(err error) error {
	var pqErr *pq.Error

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return app2.ErrNotFound
	case errors.As(err, &pqErr):
		return constraint(err.(*pq.Error))
	default:
		return err
	}
}

func constraint(pqErr *pq.Error) error {
	switch {
	case strings.HasSuffix(pqErr.Message, fmt.Sprintf("unique constraint \"%s\"", duplEmail)):
		return app2.ErrEmailExist
	case strings.HasSuffix(pqErr.Message, fmt.Sprintf("unique constraint \"%s\"", duplUsername)):
		return app2.ErrUsernameExist
	default:
		return pqErr
	}
}
