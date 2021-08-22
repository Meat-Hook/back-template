package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	unautnError "github.com/go-openapi/errors"

	"github.com/Meat-Hook/back-template/cmd/user/internal/app"
)

const (
	cookieTokenName = "authKey"
)

func (s *service) cookieKeyAuth(ctx context.Context, raw string) (*app.Session, error) {
	session, err := s.app.Auth(ctx, parseToken(raw))
	switch {
	case errors.Is(err, app.ErrNotFound):
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
