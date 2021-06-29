package web_test

import (
	"testing"

	web2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/api/web"
	operations2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/api/web/generated/client/operations"
	models2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/api/web/generated/models"
	app2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/app"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
)

func TestService_VerificationEmail(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		email  string
		appErr error
		want   *models2.Error
	}{
		"success":         {"notExist@mail.com", nil, nil},
		"err_email_exist": {"email@mail.com", app2.ErrEmailExist, APIError(app2.ErrEmailExist.Error())},
		"err_any":         {"email@mail.com", errAny, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, mockApp, client, assert := start(t)

			mockApp.EXPECT().VerificationEmail(gomock.Any(), tc.email).Return(tc.appErr)

			email := models2.Email(tc.email)
			params := operations2.NewVerificationEmailParams().
				WithArgs(operations2.VerificationEmailBody{Email: &email})
			_, err := client.Operations.VerificationEmail(params)
			assert.Equal(tc.want, errPayload(err))
		})
	}
}

func TestService_VerificationUsername(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		username string
		appErr   error
		want     *models2.Error
	}{
		"success":            {"freeUsername", nil, nil},
		"err_username_exist": {"existUsername", app2.ErrUsernameExist, APIError(app2.ErrUsernameExist.Error())},
		"err_any":            {"existUsername", errAny, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, mockApp, client, assert := start(t)

			mockApp.EXPECT().VerificationUsername(gomock.Any(), tc.username).Return(tc.appErr)

			username := models2.Username(tc.username)
			params := operations2.NewVerificationUsernameParams().
				WithArgs(operations2.VerificationUsernameBody{Username: &username})
			_, err := client.Operations.VerificationUsername(params)
			assert.Equal(tc.want, errPayload(err))
		})
	}
}

func TestService_CreateUser(t *testing.T) {
	t.Parallel()

	const (
		username = `user`
		email    = `email@mail.com`
		pass     = `password`
	)

	uid := uuid.Must(uuid.NewV4())

	testCases := map[string]struct {
		id      uuid.UUID
		appErr  error
		want    *operations2.CreateUserOK
		wantErr *models2.Error
	}{
		"success":            {uid, nil, &operations2.CreateUserOK{Payload: &operations2.CreateUserOKBody{ID: models2.UserID(uid.String())}}, nil},
		"err_email_exist":    {uuid.Nil, app2.ErrEmailExist, nil, APIError(app2.ErrEmailExist.Error())},
		"err_username_exist": {uuid.Nil, app2.ErrUsernameExist, nil, APIError(app2.ErrUsernameExist.Error())},
		"err_any":            {uuid.Nil, errAny, nil, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, mockApp, client, assert := start(t)

			mockApp.EXPECT().
				CreateUser(gomock.Any(), email, username, pass).
				Return(tc.id, tc.appErr)

			email := models2.Email(email)
			pass := models2.Password(pass)
			username := models2.Username(username)
			params := operations2.NewCreateUserParams().WithArgs(&models2.CreateUserParams{
				Email:    &email,
				Password: &pass,
				Username: &username,
			})

			res, err := client.Operations.CreateUser(params)
			assert.Equal(tc.wantErr, errPayload(err))
			assert.Equal(tc.want, res)
		})
	}
}

func TestService_GetUser(t *testing.T) {
	t.Parallel()

	restUser := web2.User(&user)
	testCases := map[string]struct {
		arg     uuid.UUID
		user    *app2.User
		appErr  error
		want    *operations2.GetUserOK
		wantErr *models2.Error
	}{
		"success":       {user.ID, &user, nil, &operations2.GetUserOK{Payload: restUser}, nil},
		"err_not_found": {uuid.Must(uuid.NewV4()), nil, app2.ErrNotFound, nil, APIError(app2.ErrNotFound.Error())},
		"err_any":       {uuid.Must(uuid.NewV4()), nil, errAny, nil, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, mockApp, client, assert := start(t)

			mockApp.EXPECT().UserByID(gomock.Any(), session, tc.arg).Return(tc.user, tc.appErr)
			mockApp.EXPECT().Auth(gomock.Any(), token).Return(&session, nil)

			uid := strfmt.UUID(tc.arg.String())

			params := operations2.NewGetUserParams().WithID(&uid)
			res, err := client.Operations.GetUser(params, apiKeyAuth)
			assert.Equal(tc.wantErr, errPayload(err))
			assert.Equal(tc.want, res)
		})
	}
}

func TestService_DeleteUser(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		appErr error
		want   *models2.Error
	}{
		"success": {nil, nil},
		"err_any": {errAny, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, mockApp, client, assert := start(t)

			mockApp.EXPECT().DeleteUser(gomock.Any(), session).Return(tc.appErr)
			mockApp.EXPECT().Auth(gomock.Any(), token).Return(&session, nil)

			params := operations2.NewDeleteUserParams()
			_, err := client.Operations.DeleteUser(params, apiKeyAuth)
			assert.Equal(tc.want, errPayload(err))
		})
	}
}

func TestServiceUpdatePassword(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		oldPass, newPass string
		appErr           error
		want             *models2.Error
	}{
		"success":                {"old_pass", "NewPassword", nil, nil},
		"err_not_valid_password": {"notCorrectPass", "NewPassword", app2.ErrNotValidPassword, APIError(app2.ErrNotValidPassword.Error())},
		"err_any":                {"notCorrectPass2", "NewPassword", errAny, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, mockApp, client, assert := start(t)

			mockApp.EXPECT().UpdatePassword(gomock.Any(), session, tc.oldPass, tc.newPass).Return(tc.appErr)
			mockApp.EXPECT().Auth(gomock.Any(), token).Return(&session, nil)

			newPass := models2.Password(tc.newPass)
			lastPass := models2.Password(tc.oldPass)
			params := operations2.NewUpdatePasswordParams().WithArgs(&models2.UpdatePassword{
				New: &newPass,
				Old: &lastPass,
			})
			_, err := client.Operations.UpdatePassword(params, apiKeyAuth)
			assert.Equal(tc.want, errPayload(err))
		})
	}
}

func TestServiceUpdateUsername(t *testing.T) {
	t.Parallel()

	const userName = `zergsLaw`

	testCases := map[string]struct {
		appErr error
		want   *models2.Error
	}{
		"success":                    {nil, nil},
		"err_username_exist":         {app2.ErrUsernameExist, APIError(app2.ErrUsernameExist.Error())},
		"err_username_not_different": {app2.ErrNotDifferent, APIError(app2.ErrNotDifferent.Error())},
		"err_any":                    {errAny, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, mockApp, client, assert := start(t)

			mockApp.EXPECT().UpdateUsername(gomock.Any(), session, userName).Return(tc.appErr)
			mockApp.EXPECT().Auth(gomock.Any(), token).Return(&session, nil)

			userName := models2.Username(userName)
			params := operations2.NewUpdateUsernameParams().
				WithArgs(operations2.UpdateUsernameBody{Username: &userName})

			_, err := client.Operations.UpdateUsername(params, apiKeyAuth)
			assert.Equal(tc.want, errPayload(err))
		})
	}
}

func TestServiceGetUsers(t *testing.T) {
	t.Parallel()

	const userName = `zergsL`

	testCases := map[string]struct {
		users     []app2.User
		appErr    error
		want      *operations2.GetUsersOK
		wantTotal int32
		wantErr   *models2.Error
	}{
		"success": {[]app2.User{user}, nil, &operations2.GetUsersOK{Payload: &operations2.GetUsersOKBody{Total: swag.Int32(1), Users: web2.Users([]app2.User{user})}}, 1, nil},
		"err_any": {nil, errAny, nil, 0, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, mockApp, client, assert := start(t)

			mockApp.EXPECT().
				ListUserByUsername(gomock.Any(), session, userName, app2.SearchParams{Limit: 10}).
				Return(tc.users, len(tc.users), tc.appErr)
			mockApp.EXPECT().Auth(gomock.Any(), token).Return(&session, nil)

			params := operations2.NewGetUsersParams().
				WithLimit(10).
				WithOffset(swag.Int32(0)).
				WithUsername(userName)

			res, err := client.Operations.GetUsers(params, apiKeyAuth)
			assert.Equal(tc.wantErr, errPayload(err))
			assert.Equal(tc.want, res)
		})
	}
}
