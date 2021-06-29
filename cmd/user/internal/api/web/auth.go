package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	app2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/app"
	unautnError "github.com/go-openapi/errors"
)

const (
	cookieTokenName = "authKey"
	authTimeout     = 250 * time.Millisecond
)

func (svc *service) cookieKeyAuth(raw string) (*app2.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), authTimeout)
	defer cancel()

	session, err := svc.app.Auth(ctx, parseToken(raw))
	switch {
	case errors.Is(err, app2.ErrNotFound):
		return nil, unautnError.Unauthenticated("user")
	case err != nil:
		return nil, fmt.Errorf("auth: %w", err)
	default:
		return session, nil
	}
}

func parseToken(raw string) string {
	header := http.Header{}
	header.Add("Cookie", raw)
	request := http.Request{Header: header}
	cookieKey, err := request.Cookie(cookieTokenName)
	if err != nil {
		return ""
	}

	return cookieKey.Value
}
