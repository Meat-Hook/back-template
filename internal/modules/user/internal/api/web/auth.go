package web

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Meat-Hook/back-template/internal/modules/user/internal/app"
	unautnError "github.com/go-openapi/errors"
)

const (
	authTimeout = 250 * time.Millisecond
)

func (svc *service) cookieKeyAuth(raw string) (*app.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), authTimeout)
	defer cancel()
	session, err := svc.app.Auth(ctx, raw)
	switch {
	case errors.Is(err, app.ErrNotFound):
		return nil, unautnError.Unauthenticated("user")
	case err != nil:
		return nil, fmt.Errorf("auth: %w", err)
	default:
		return session, nil
	}
}


