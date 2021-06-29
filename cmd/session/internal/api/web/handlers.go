package web

import (
	"errors"
	"net/http"

	models2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/api/web/generated/models"
	operations2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/api/web/generated/restapi/operations"
	app2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/app"
	"github.com/go-openapi/swag"
	"github.com/rs/zerolog"
)

func (svc *service) login(params operations2.LoginParams) operations2.LoginResponder {
	ctx, log, remoteIP := fromRequest(params.HTTPRequest, nil)

	origin := app2.Origin{
		IP:        remoteIP,
		UserAgent: params.HTTPRequest.Header.Get("User-Agent"),
	}

	u, token, err := svc.app.Login(ctx, string(*params.Args.Email), string(*params.Args.Password), origin)
	defer logs(log, err)
	switch {
	case err == nil:
		return operations2.NewLoginOK().WithPayload(User(u)).
			WithSetCookie(generateCookie(token.Value).String())
	case errors.Is(err, app2.ErrNotFound):
		return operations2.NewLoginDefault(http.StatusNotFound).WithPayload(apiError(app2.ErrNotFound.Error()))
	case errors.Is(err, app2.ErrNotValidPassword):
		return operations2.NewLoginDefault(http.StatusBadRequest).WithPayload(apiError(app2.ErrNotValidPassword.Error()))
	default:
		return operations2.NewLoginDefault(http.StatusInternalServerError).
			WithPayload(apiError(http.StatusText(http.StatusInternalServerError)))
	}
}

func (svc *service) logout(params operations2.LogoutParams, session *app2.Session) operations2.LogoutResponder {
	ctx, log, _ := fromRequest(params.HTTPRequest, session)

	err := svc.app.Logout(ctx, *session)
	defer logs(log, err)
	switch {
	case err == nil:
		return operations2.NewLogoutNoContent()
	default:
		return operations2.NewLogoutDefault(http.StatusInternalServerError).
			WithPayload(apiError(http.StatusText(http.StatusInternalServerError)))
	}
}

// User conversion app.User => models.User.
func User(u *app2.User) *models2.User {
	id := models2.UserID(u.ID.String())
	username := models2.Username(u.Name)
	email := models2.Email(u.Email)

	return &models2.User{
		ID:       &id,
		Username: &username,
		Email:    &email,
	}
}

func apiError(txt string) *models2.Error {
	return &models2.Error{
		Message: swag.String(txt),
	}
}

func logs(log zerolog.Logger, err error) {
	if err != nil {
		log.Error().Err(err).Send()
	}
}
