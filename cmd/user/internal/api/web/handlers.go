package web

import (
	"errors"
	"net/http"

	models2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/api/web/generated/models"
	operations2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/api/web/generated/restapi/operations"
	app2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/app"
	"github.com/go-openapi/swag"
	"github.com/gofrs/uuid"
)

func (svc *service) verificationEmail(params operations2.VerificationEmailParams) operations2.VerificationEmailResponder {
	ctx, log := fromRequest(params.HTTPRequest, nil)

	err := svc.app.VerificationEmail(ctx, string(*params.Args.Email))
	defer logs(log, err)
	switch {
	case err == nil:
		return operations2.NewVerificationEmailNoContent()
	case errors.Is(err, app2.ErrEmailExist):
		return operations2.NewVerificationEmailDefault(http.StatusConflict).WithPayload(apiError(app2.ErrEmailExist.Error()))
	default:
		return operations2.NewVerificationEmailDefault(http.StatusInternalServerError).
			WithPayload(apiError(http.StatusText(http.StatusInternalServerError)))
	}
}

func (svc *service) verificationUsername(params operations2.VerificationUsernameParams) operations2.VerificationUsernameResponder {
	ctx, log := fromRequest(params.HTTPRequest, nil)

	err := svc.app.VerificationUsername(ctx, string(*params.Args.Username))
	defer logs(log, err)
	switch {
	case err == nil:
		return operations2.NewVerificationUsernameNoContent()
	case errors.Is(err, app2.ErrUsernameExist):
		return operations2.NewVerificationUsernameDefault(http.StatusConflict).WithPayload(apiError(app2.ErrUsernameExist.Error()))
	default:
		return operations2.NewVerificationUsernameDefault(http.StatusInternalServerError).
			WithPayload(apiError(http.StatusText(http.StatusInternalServerError)))
	}
}

func (svc *service) createUser(params operations2.CreateUserParams) operations2.CreateUserResponder {
	ctx, log := fromRequest(params.HTTPRequest, nil)

	id, err := svc.app.CreateUser(
		ctx,
		string(*params.Args.Email),
		string(*params.Args.Username),
		string(*params.Args.Password),
	)
	defer logs(log, err)
	switch {
	case err == nil:
		return operations2.NewCreateUserOK().WithPayload(&operations2.CreateUserOKBody{ID: models2.UserID(id.String())})
	case errors.Is(err, app2.ErrEmailExist):
		return operations2.NewCreateUserDefault(http.StatusConflict).WithPayload(apiError(app2.ErrEmailExist.Error()))
	case errors.Is(err, app2.ErrUsernameExist):
		return operations2.NewCreateUserDefault(http.StatusConflict).WithPayload(apiError(app2.ErrUsernameExist.Error()))
	default:
		return operations2.NewCreateUserDefault(http.StatusInternalServerError).
			WithPayload(apiError(http.StatusText(http.StatusInternalServerError)))
	}
}

func (svc *service) getUser(params operations2.GetUserParams, session *app2.Session) operations2.GetUserResponder {
	ctx, log := fromRequest(params.HTTPRequest, session)

	getUserID := session.UserID
	if params.ID != nil {
		getUserID = uuid.FromStringOrNil(params.ID.String())
	}

	u, err := svc.app.UserByID(ctx, *session, getUserID)
	defer logs(log, err)
	switch {
	case err == nil:
		return operations2.NewGetUserOK().WithPayload(User(u))
	case errors.Is(err, app2.ErrNotFound):
		return operations2.NewGetUserDefault(http.StatusNotFound).WithPayload(apiError(app2.ErrNotFound.Error()))
	default:
		return operations2.NewGetUserDefault(http.StatusInternalServerError).
			WithPayload(apiError(http.StatusText(http.StatusInternalServerError)))
	}
}

func (svc *service) deleteUser(params operations2.DeleteUserParams, session *app2.Session) operations2.DeleteUserResponder {
	ctx, log := fromRequest(params.HTTPRequest, session)

	err := svc.app.DeleteUser(ctx, *session)
	defer logs(log, err)
	switch {
	case err == nil:
		return operations2.NewDeleteUserNoContent()
	default:
		return operations2.NewDeleteUserDefault(http.StatusInternalServerError).
			WithPayload(apiError(http.StatusText(http.StatusInternalServerError)))
	}
}

func (svc *service) updatePassword(params operations2.UpdatePasswordParams, session *app2.Session) operations2.UpdatePasswordResponder {
	ctx, log := fromRequest(params.HTTPRequest, session)

	err := svc.app.UpdatePassword(ctx, *session, string(*params.Args.Old), string(*params.Args.New))
	defer logs(log, err)
	switch {
	case err == nil:
		return operations2.NewUpdatePasswordNoContent()
	case errors.Is(err, app2.ErrNotValidPassword):
		return operations2.NewUpdatePasswordDefault(http.StatusBadRequest).
			WithPayload(apiError(app2.ErrNotValidPassword.Error()))
	default:
		return operations2.NewUpdatePasswordDefault(http.StatusInternalServerError).
			WithPayload(apiError(http.StatusText(http.StatusInternalServerError)))
	}
}

func (svc *service) updateUsername(params operations2.UpdateUsernameParams, session *app2.Session) operations2.UpdateUsernameResponder {
	ctx, log := fromRequest(params.HTTPRequest, session)

	err := svc.app.UpdateUsername(ctx, *session, string(*params.Args.Username))
	defer logs(log, err)
	switch {
	case err == nil:
		return operations2.NewUpdateUsernameNoContent()
	case errors.Is(err, app2.ErrUsernameExist):
		return operations2.NewUpdateUsernameDefault(http.StatusConflict).
			WithPayload(apiError(app2.ErrUsernameExist.Error()))
	case errors.Is(err, app2.ErrNotDifferent):
		return operations2.NewUpdateUsernameDefault(http.StatusConflict).
			WithPayload(apiError(app2.ErrNotDifferent.Error()))
	default:
		return operations2.NewUpdateUsernameDefault(http.StatusInternalServerError).
			WithPayload(apiError(http.StatusText(http.StatusInternalServerError)))
	}
}

func (svc *service) getUsers(params operations2.GetUsersParams, session *app2.Session) operations2.GetUsersResponder {
	ctx, log := fromRequest(params.HTTPRequest, session)

	page := app2.SearchParams{
		Limit:  uint(params.Limit),
		Offset: uint(swag.Int32Value(params.Offset)),
	}

	u, total, err := svc.app.ListUserByUsername(ctx, *session, params.Username, page)
	defer logs(log, err)
	switch {
	case err == nil:
		return operations2.NewGetUsersOK().WithPayload(&operations2.GetUsersOKBody{
			Total: swag.Int32(int32(total)),
			Users: Users(u),
		})
	default:
		return operations2.NewGetUsersDefault(http.StatusInternalServerError).
			WithPayload(apiError(http.StatusText(http.StatusInternalServerError)))
	}
}
