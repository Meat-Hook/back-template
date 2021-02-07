package web

import (
	"errors"
	"net/http"

	"github.com/Meat-Hook/back-template/internal/modules/session/internal/api/web/generated/models"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/api/web/generated/restapi/operations"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/app"
	"github.com/go-openapi/swag"
	"github.com/rs/zerolog"
)

func (svc *service) login(params operations.LoginParams) operations.LoginResponder {
	ctx, log, remoteIP := fromRequest(params.HTTPRequest, nil)

	origin := app.Origin{
		IP:        remoteIP,
		UserAgent: params.HTTPRequest.Header.Get("User-Agent"),
	}

	u, token, err := svc.app.Login(ctx, string(params.Args.Email), string(params.Args.Password), origin)
	defer logs(log, err)
	switch {
	case err == nil:
		return operations.NewLoginOK().WithPayload(User(u)).
			WithSetCookie(generateCookie(token.Value).String())
	case errors.Is(err, app.ErrNotFound):
		return operations.NewLoginDefault(http.StatusNotFound).WithPayload(apiError(err.Error()))
	case errors.Is(err, app.ErrNotValidPassword):
		return operations.NewLoginDefault(http.StatusBadRequest).WithPayload(apiError(err.Error()))
	default:
		return operations.NewLoginDefault(http.StatusInternalServerError).
			WithPayload(apiError(http.StatusText(http.StatusInternalServerError)))
	}
}

func (svc *service) logout(params operations.LogoutParams, session *app.Session) operations.LogoutResponder {
	ctx, log, _ := fromRequest(params.HTTPRequest, session)

	err := svc.app.Logout(ctx, *session)
	defer logs(log, err)
	switch {
	case err == nil:
		return operations.NewLogoutNoContent()
	default:
		return operations.NewLogoutDefault(http.StatusInternalServerError).
			WithPayload(apiError(http.StatusText(http.StatusInternalServerError)))
	}
}

// User conversion app.User => models.User.
func User(u *app.User) *models.User {
	return &models.User{
		ID:       models.UserID(u.ID),
		Username: models.Username(u.Name),
		Email:    models.Email(u.Email),
	}
}

func apiError(txt string) *models.Error {
	return &models.Error{
		Message: swag.String(txt),
	}
}

func logs(log zerolog.Logger, err error) {
	if err != nil {
		log.Error().Err(err).Send()
	}
}
