package web_test

import (
	"testing"

	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/web"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/web/generated/client/operations"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/web/generated/models"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/app"
	"github.com/go-openapi/swag"
	"github.com/golang/mock/gomock"
)

func TestService_VerificationEmail(t *testing.T) {
	t.Parallel()

	_, mockApp, client, assert := start(t)

	testCases := map[string]struct {
		email  string
		appErr error
		want   *models.Error
	}{
		"success":   {"notExist@mail.com", nil, nil},
		"exist":     {"email@mail.com", app.ErrEmailExist, APIError(app.ErrEmailExist.Error())},
		"any_error": {"email@mail.com", errAny, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			mockApp.EXPECT().VerificationEmail(gomock.Any(), tc.email).Return(tc.appErr)

			params := operations.NewVerificationEmailParams().
				WithArgs(operations.VerificationEmailBody{Email: models.Email(tc.email)})
			_, err := client.Operations.VerificationEmail(params)
			assert.Equal(tc.want, errPayload(err))
		})
	}
}

func TestService_VerificationUsername(t *testing.T) {
	t.Parallel()

	_, mockApp, client, assert := start(t)

	testCases := map[string]struct {
		username string
		appErr   error
		want     *models.Error
	}{
		"success":   {"freeUsername", nil, nil},
		"exist":     {"existUsername", app.ErrUsernameExist, APIError(app.ErrUsernameExist.Error())},
		"any_error": {"existUsername", errAny, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			mockApp.EXPECT().VerificationUsername(gomock.Any(), tc.username).Return(tc.appErr)

			params := operations.NewVerificationUsernameParams().
				WithArgs(operations.VerificationUsernameBody{Username: models.Username(tc.username)})
			_, err := client.Operations.VerificationUsername(params)
			assert.Equal(tc.want, errPayload(err))
		})
	}
}

func TestService_CreateUser(t *testing.T) {
	t.Parallel()

	_, mockApp, client, assert := start(t)

	const (
		username = `user`
		email    = `email@mail.com`
		pass     = `password`
	)

	testCases := map[string]struct {
		id      int
		appErr  error
		want    *operations.CreateUserOKBody
		wantErr *models.Error
	}{
		"success":        {1, nil, &operations.CreateUserOKBody{ID: 1}, nil},
		"email_exist":    {0, app.ErrEmailExist, nil, APIError(app.ErrEmailExist.Error())},
		"username_exist": {0, app.ErrUsernameExist, nil, APIError(app.ErrUsernameExist.Error())},
		"internal_error": {0, errAny, nil, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			mockApp.EXPECT().
				CreateUser(gomock.Any(), email, username, pass).
				Return(tc.id, tc.appErr)

			params := operations.NewCreateUserParams().WithArgs(&models.CreateUserParams{
				Email:    email,
				Password: pass,
				Username: username,
			})

			res, err := client.Operations.CreateUser(params)
			if tc.wantErr == nil {
				assert.Nil(err)
				assert.Equal(tc.want, res.Payload)
			} else {
				assert.Nil(res)
				assert.Equal(tc.wantErr, errPayload(err))
			}
		})
	}
}

func TestService_GetUser(t *testing.T) {
	t.Parallel()

	_, mockApp, client, assert := start(t)

	restUser := web.User(&user)
	testCases := map[string]struct {
		arg     int
		user    *app.User
		appErr  error
		want    *models.User
		wantErr *models.Error
	}{
		"success":   {user.ID, &user, nil, restUser, nil},
		"not_found": {2, nil, app.ErrNotFound, nil, APIError(app.ErrNotFound.Error())},
		"any_error": {3, nil, errAny, nil, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			mockApp.EXPECT().UserByID(gomock.Any(), session, tc.arg).Return(tc.user, tc.appErr)
			mockApp.EXPECT().Auth(gomock.Any(), token).Return(&session, nil)

			params := operations.NewGetUserParams().WithID(swag.Int64(int64(tc.arg)))
			res, err := client.Operations.GetUser(params, apiKeyAuth)
			if tc.wantErr == nil {
				assert.Nil(err)
				assert.Equal(tc.want, res.Payload)
			} else {
				assert.Nil(res)
				assert.Equal(tc.wantErr, errPayload(err))
			}
		})
	}
}

func TestService_DeleteUser(t *testing.T) {
	t.Parallel()

	_, mockApp, client, assert := start(t)

	testCases := map[string]struct {
		appErr error
		want   *models.Error
	}{
		"success":   {nil, nil},
		"any_error": {errAny, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			mockApp.EXPECT().DeleteUser(gomock.Any(), session).Return(tc.appErr)
			mockApp.EXPECT().Auth(gomock.Any(), token).Return(&session, nil)

			params := operations.NewDeleteUserParams()
			_, err := client.Operations.DeleteUser(params, apiKeyAuth)
			assert.Equal(tc.want, errPayload(err))
		})
	}
}

func TestServiceUpdatePassword(t *testing.T) {
	t.Parallel()

	_, mockApp, client, assert := start(t)

	testCases := map[string]struct {
		oldPass, newPass string
		appErr           error
		want             *models.Error
	}{
		"success":            {"old_pass", "NewPassword", nil, nil},
		"not_valid_password": {"notCorrectPass", "NewPassword", app.ErrNotValidPassword, APIError(app.ErrNotValidPassword.Error())},
		"any_error":          {"notCorrectPass2", "NewPassword", errAny, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			mockApp.EXPECT().UpdatePassword(gomock.Any(), session, tc.oldPass, tc.newPass).Return(tc.appErr)
			mockApp.EXPECT().Auth(gomock.Any(), token).Return(&session, nil)

			params := operations.NewUpdatePasswordParams().WithArgs(&models.UpdatePassword{
				New: models.Password(tc.newPass),
				Old: models.Password(tc.oldPass),
			})
			_, err := client.Operations.UpdatePassword(params, apiKeyAuth)
			assert.Equal(tc.want, errPayload(err))
		})
	}
}

func TestServiceUpdateUsername(t *testing.T) {
	t.Parallel()

	_, mockApp, client, assert := start(t)

	const userName = `zergsLaw`

	testCases := map[string]struct {
		appErr error
		want   *models.Error
	}{
		"success":                {nil, nil},
		"username_exist":         {app.ErrUsernameExist, APIError(app.ErrUsernameExist.Error())},
		"username_not_different": {app.ErrNotDifferent, APIError(app.ErrNotDifferent.Error())},
		"any_error":              {errAny, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			mockApp.EXPECT().UpdateUsername(gomock.Any(), session, userName).Return(tc.appErr)
			mockApp.EXPECT().Auth(gomock.Any(), token).Return(&session, nil)

			params := operations.NewUpdateUsernameParams().
				WithArgs(operations.UpdateUsernameBody{Username: userName})

			_, err := client.Operations.UpdateUsername(params, apiKeyAuth)
			assert.Equal(tc.want, errPayload(err))
		})
	}
}

func TestServiceGetUsers(t *testing.T) {
	t.Parallel()

	_, mockApp, client, assert := start(t)

	const userName = `zergsL`

	testCases := map[string]struct {
		users     []app.User
		appErr    error
		want      []*models.User
		wantTotal int32
		wantErr   *models.Error
	}{
		"success":   {[]app.User{user}, nil, web.Users([]app.User{user}), 1, nil},
		"any_error": {nil, errAny, nil, 0, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			mockApp.EXPECT().
				ListUserByUsername(gomock.Any(), session, userName, app.SearchParams{Limit: 10}).
				Return(tc.users, len(tc.users), tc.appErr)
			mockApp.EXPECT().Auth(gomock.Any(), token).Return(&session, nil)

			params := operations.NewGetUsersParams().
				WithLimit(10).
				WithOffset(swag.Int32(0)).
				WithUsername(userName)

			res, err := client.Operations.GetUsers(params, apiKeyAuth)
			if tc.wantErr == nil {
				assert.Nil(err)
				assert.Equal(&operations.GetUsersOKBody{
					Total: swag.Int32(int32(len(tc.users))),
					Users: tc.want,
				}, res.Payload)
			} else {
				assert.Nil(res)
				assert.Equal(tc.wantErr, errPayload(err))
			}
		})
	}
}
