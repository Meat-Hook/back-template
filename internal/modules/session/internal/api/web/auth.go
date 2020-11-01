package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Meat-Hook/back-template/internal/modules/session/internal/app"
	unautnError "github.com/go-openapi/errors"
)

const (
	cookieTokenName = "authKey"
	authTimeout     = 250 * time.Millisecond
)

func (svc *service) cookieKeyAuth(raw string) (*app.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), authTimeout)
	defer cancel()

	session, err := svc.app.Session(ctx, parseToken(raw, cookieTokenName))
	switch {
	case errors.Is(err, app.ErrNotFound):
		return nil, unautnError.Unauthenticated("session")
	case err != nil:
		return nil, fmt.Errorf("auth: %w", err)
	default:
		return session, nil
	}
}

func parseToken(raw string, name string) string {
	header := http.Header{}
	header.Add("Cookie", raw)
	request := http.Request{Header: header} // nolint:exhaustivestruct
	cookieKey, err := request.Cookie(name)
	if err != nil {
		return ""
	}

	return cookieKey.Value
}

func generateCookie(token string) *http.Cookie {
	cookie := &http.Cookie{
		Name:       cookieTokenName,
		Value:      token,
		Path:       "/",
		Domain:     "",
		Expires:    time.Time{},
		RawExpires: "",
		MaxAge:     0,
		Secure:     true,
		HttpOnly:   true,
		SameSite:   http.SameSiteLaxMode,
		Raw:        "",
		Unparsed:   nil,
	}

	return cookie
}
