package web

import (
	"errors"
	"net/http"

	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/web/generated/models"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/web/generated/restapi/operations"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/app"
	"github.com/go-openapi/swag"
)

func (svc *service) verificationEmail(params operations.VerificationEmailParams) operations.VerificationEmailResponder {
	ctx, log := fromRequest(params.HTTPRequest, nil)

	err := svc.app.VerificationEmail(ctx, string(params.Args.Email))
	defer logs(log, err)
	switch {
	case err == nil:
		return operations.NewVerificationEmailNoContent()
	case errors.Is(err, app.ErrEmailExist):
		return operations.NewVerificationEmailDefault(http.StatusConflict).WithPayload(apiError(err.Error()))
	default:
		return operations.NewVerificationEmailDefault(http.StatusInternalServerError).
			WithPayload(apiError(http.StatusText(http.StatusInternalServerError)))
	}
}

func (svc *service) verificationUsername(params operations.VerificationUsernameParams) operations.VerificationUsernameResponder {
	ctx, log := fromRequest(params.HTTPRequest, nil)

	err := svc.app.VerificationUsername(ctx, string(params.Args.Username))
	defer logs(log, err)
	switch {
	case err == nil:
		return operations.NewVerificationUsernameNoContent()
	case errors.Is(err, app.ErrUsernameExist):
		return operations.NewVerificationUsernameDefault(http.StatusConflict).WithPayload(apiError(err.Error()))
	default:
		return operations.NewVerificationUsernameDefault(http.StatusInternalServerError).
			WithPayload(apiError(http.StatusText(http.StatusInternalServerError)))
	}
}

func (svc *service) createUser(params operations.CreateUserParams) operations.CreateUserResponder {
	ctx, log := fromRequest(params.HTTPRequest, nil)

	id, err := svc.app.CreateUser(
		ctx,
		string(params.Args.Email),
		string(params.Args.Username),
		string(params.Args.Password),
	)
	defer logs(log, err)
	switch {
	case err == nil:
		return operations.NewCreateUserOK().WithPayload(&operations.CreateUserOKBody{ID: models.UserID(id)})
	case errors.Is(err, app.ErrEmailExist):
		return operations.NewCreateUserDefault(http.StatusConflict).WithPayload(apiError(err.Error()))
	case errors.Is(err, app.ErrUsernameExist):
		return operations.NewCreateUserDefault(http.StatusConflict).WithPayload(apiError(err.Error()))
	default:
		return operations.NewCreateUserDefault(http.StatusInternalServerError).
			WithPayload(apiError(http.StatusText(http.StatusInternalServerError)))
	}
}

func (svc *service) getUser(params operations.GetUserParams, session *app.Session) operations.GetUserResponder {
	ctx, log := fromRequest(params.HTTPRequest, session)

	getUserID := session.UserID
	if params.ID != nil {
		getUserID = int(*params.ID)
	}

	u, err := svc.app.UserByID(ctx, *session, getUserID)
	defer logs(log, err)
	switch {
	case err == nil:
		return operations.NewGetUserOK().WithPayload(User(u))
	case errors.Is(err, app.ErrNotFound):
		return operations.NewGetUserDefault(http.StatusNotFound).WithPayload(apiError(err.Error()))
	default:
		return operations.NewGetUserDefault(http.StatusInternalServerError).
			WithPayload(apiError(http.StatusText(http.StatusInternalServerError)))
	}
}

func (svc *service) deleteUser(params operations.DeleteUserParams, session *app.Session) operations.DeleteUserResponder {
	ctx, log := fromRequest(params.HTTPRequest, session)

	err := svc.app.DeleteUser(ctx, *session)
	defer logs(log, err)
	switch {
	case err == nil:
		return operations.NewDeleteUserNoContent()
	default:
		return operations.NewDeleteUserDefault(http.StatusInternalServerError).
			WithPayload(apiError(http.StatusText(http.StatusInternalServerError)))
	}
}

func (svc *service) updatePassword(params operations.UpdatePasswordParams, session *app.Session) operations.UpdatePasswordResponder {
	ctx, log := fromRequest(params.HTTPRequest, session)

	err := svc.app.UpdatePassword(ctx, *session, string(params.Args.Old), string(params.Args.New))
	defer logs(log, err)
	switch {
	case err == nil:
		return operations.NewUpdatePasswordNoContent()
	case errors.Is(err, app.ErrNotValidPassword):
		return operations.NewUpdatePasswordDefault(http.StatusBadRequest).
			WithPayload(apiError(err.Error()))
	default:
		return operations.NewUpdatePasswordDefault(http.StatusInternalServerError).
			WithPayload(apiError(http.StatusText(http.StatusInternalServerError)))
	}
}

func (svc *service) updateUsername(params operations.UpdateUsernameParams, session *app.Session) operations.UpdateUsernameResponder {
	ctx, log := fromRequest(params.HTTPRequest, session)

	err := svc.app.UpdateUsername(ctx, *session, string(params.Args.Username))
	defer logs(log, err)
	switch {
	case err == nil:
		return operations.NewUpdateUsernameNoContent()
	case errors.Is(err, app.ErrUsernameExist):
		return operations.NewUpdateUsernameDefault(http.StatusConflict).
			WithPayload(apiError(err.Error()))
	case errors.Is(err, app.ErrNotDifferent):
		return operations.NewUpdateUsernameDefault(http.StatusConflict).
			WithPayload(apiError(err.Error()))
	default:
		return operations.NewUpdateUsernameDefault(http.StatusInternalServerError).
			WithPayload(apiError(http.StatusText(http.StatusInternalServerError)))
	}
}

func (svc *service) getUsers(params operations.GetUsersParams, session *app.Session) operations.GetUsersResponder {
	ctx, log := fromRequest(params.HTTPRequest, session)

	page := app.SearchParams{
		Limit:  uint(params.Limit),
		Offset: uint(swag.Int32Value(params.Offset)),
	}

	u, total, err := svc.app.ListUserByUsername(ctx, *session, params.Username, page)
	defer logs(log, err)
	switch {
	case err == nil:
		return operations.NewGetUsersOK().WithPayload(&operations.GetUsersOKBody{
			Total: swag.Int32(int32(total)),
			Users: Users(u),
		})
	default:
		return operations.NewGetUsersDefault(http.StatusInternalServerError).
			WithPayload(apiError(http.StatusText(http.StatusInternalServerError)))
	}
}
