package web_test

import (
	"bytes"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"

	"github.com/Meat-Hook/back-template/cmd/user/internal/api/web"
	"github.com/Meat-Hook/back-template/cmd/user/internal/api/web/generated/client/operations"
	"github.com/Meat-Hook/back-template/cmd/user/internal/api/web/generated/models"
	"github.com/Meat-Hook/back-template/cmd/user/internal/app"
)

func TestService_VerificationEmail(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		email  string
		appErr error
		want   *models.Error
	}{
		"success":         {"notExist@mail.com", nil, nil},
		"err_email_exist": {"email@mail.com", app.ErrEmailExist, APIError(app.ErrEmailExist.Error())},
		"err_any":         {"email@mail.com", errAny, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, mockApp, client, assert := start(t)

			mockApp.EXPECT().VerificationEmail(gomock.Any(), tc.email).Return(tc.appErr)

			email := models.Email(tc.email)
			params := operations.NewVerificationEmailParams().
				WithArgs(operations.VerificationEmailBody{Email: &email})
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
		want     *models.Error
	}{
		"success":            {"freeUsername", nil, nil},
		"err_username_exist": {"existUsername", app.ErrUsernameExist, APIError(app.ErrUsernameExist.Error())},
		"err_any":            {"existUsername", errAny, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, mockApp, client, assert := start(t)

			mockApp.EXPECT().VerificationUsername(gomock.Any(), tc.username).Return(tc.appErr)

			username := models.Username(tc.username)
			params := operations.NewVerificationUsernameParams().
				WithArgs(operations.VerificationUsernameBody{Username: &username})
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
		want    *operations.CreateUserOK
		wantErr *models.Error
	}{
		"success":            {uid, nil, &operations.CreateUserOK{Payload: &operations.CreateUserOKBody{ID: models.UserID(uid.String())}}, nil},
		"err_email_exist":    {uuid.Nil, app.ErrEmailExist, nil, APIError(app.ErrEmailExist.Error())},
		"err_username_exist": {uuid.Nil, app.ErrUsernameExist, nil, APIError(app.ErrUsernameExist.Error())},
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

			email := models.Email(email)
			pass := models.Password(pass)
			username := models.Username(username)
			params := operations.NewCreateUserParams().WithArgs(&models.CreateUserParams{
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

	restUser := web.User(&user)
	testCases := map[string]struct {
		arg     uuid.UUID
		user    *app.User
		appErr  error
		want    *operations.GetUserOK
		wantErr *models.Error
	}{
		"success":       {user.ID, &user, nil, &operations.GetUserOK{Payload: restUser}, nil},
		"err_not_found": {uuid.Must(uuid.NewV4()), nil, app.ErrNotFound, nil, APIError(app.ErrNotFound.Error())},
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

			params := operations.NewGetUserParams().WithID(&uid)
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
		want   *models.Error
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

			params := operations.NewDeleteUserParams()
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
		want             *models.Error
	}{
		"success":                {"old_pass", "NewPassword", nil, nil},
		"err_not_valid_password": {"notCorrectPass", "NewPassword", app.ErrNotValidPassword, APIError(app.ErrNotValidPassword.Error())},
		"err_any":                {"notCorrectPass2", "NewPassword", errAny, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, mockApp, client, assert := start(t)

			mockApp.EXPECT().UpdatePassword(gomock.Any(), session, tc.oldPass, tc.newPass).Return(tc.appErr)
			mockApp.EXPECT().Auth(gomock.Any(), token).Return(&session, nil)

			newPass := models.Password(tc.newPass)
			lastPass := models.Password(tc.oldPass)
			params := operations.NewUpdatePasswordParams().WithArgs(&models.UpdatePassword{
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
		want   *models.Error
	}{
		"success":                    {nil, nil},
		"err_username_exist":         {app.ErrUsernameExist, APIError(app.ErrUsernameExist.Error())},
		"err_username_not_different": {app.ErrNotDifferent, APIError(app.ErrNotDifferent.Error())},
		"err_any":                    {errAny, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, mockApp, client, assert := start(t)

			mockApp.EXPECT().UpdateUsername(gomock.Any(), session, userName).Return(tc.appErr)
			mockApp.EXPECT().Auth(gomock.Any(), token).Return(&session, nil)

			userName := models.Username(userName)
			params := operations.NewUpdateUsernameParams().
				WithArgs(operations.UpdateUsernameBody{Username: &userName})

			_, err := client.Operations.UpdateUsername(params, apiKeyAuth)
			assert.Equal(tc.want, errPayload(err))
		})
	}
}

func TestServiceGetUsers(t *testing.T) {
	t.Parallel()

	const userName = `zergsL`

	testCases := map[string]struct {
		users     []app.User
		appErr    error
		want      *operations.GetUsersOK
		wantTotal int32
		wantErr   *models.Error
	}{
		"success": {[]app.User{user}, nil, &operations.GetUsersOK{Payload: &operations.GetUsersOKBody{Total: swag.Int32(1), Users: web.Users([]app.User{user})}}, 1, nil},
		"err_any": {nil, errAny, nil, 0, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, mockApp, client, assert := start(t)

			mockApp.EXPECT().
				ListUserByUsername(gomock.Any(), session, userName, app.SearchParams{Limit: 10}).
				Return(tc.users, len(tc.users), tc.appErr)
			mockApp.EXPECT().Auth(gomock.Any(), token).Return(&session, nil)

			params := operations.NewGetUsersParams().
				WithLimit(10).
				WithOffset(swag.Int32(0)).
				WithUsername(userName)

			res, err := client.Operations.GetUsers(params, apiKeyAuth)
			assert.Equal(tc.wantErr, errPayload(err))
			assert.Equal(tc.want, res)
		})
	}
}

func TestService_Login(t *testing.T) {
	t.Parallel()

	var (
		token = app.Token{
			Value: "token",
		}
	)

	testCases := map[string]struct {
		email, pass string
		token       *app.Token
		appErr      error
		wantErr     *models.Error
	}{
		"success":                {user.Email, "password", &token, nil, nil},
		"err_not_found":          {"notExist@email.com", "password", nil, app.ErrNotFound, APIError(app.ErrNotFound.Error())},
		"err_not_valid_password": {user.Email, "notValidPass", nil, app.ErrNotValidPassword, APIError(app.ErrNotValidPassword.Error())},
		"err_any":                {"randomEmail@email.com", "notValidPass", nil, errAny, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, mockApp, client, assert := start(t)

			mockApp.EXPECT().Login(gomock.Any(), tc.email, tc.pass, gomock.Any()).Return(tc.token, tc.appErr)

			email := models.Email(tc.email)
			password := models.Password(tc.pass)

			params := operations.NewLoginParams().
				WithArgs(&models.LoginParam{
					Email:    &email,
					Password: &password,
				})
			_, err := client.Operations.Login(params)
			assert.Equal(tc.wantErr, errPayload(err))
		})
	}
}

func TestService_Logout(t *testing.T) {
	t.Parallel()

	session := app.Session{
		ID:     uuid.Must(uuid.NewV4()),
		UserID: user.ID,
	}

	testCases := []struct {
		name   string
		appErr error
		want   *models.Error
	}{
		{"success", nil, nil},
		{"err_any", errAny, APIError("Internal Server Error")},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, mockApp, client, assert := start(t)

			mockApp.EXPECT().Logout(gomock.Any(), session).Return(tc.appErr)
			mockApp.EXPECT().Auth(gomock.Any(), token).Return(&session, nil)

			params := operations.NewLogoutParams()
			_, err := client.Operations.Logout(params, apiKeyAuth)
			assert.Equal(tc.want, errPayload(err))
		})
	}
}

func TestService_UploadAvatar(t *testing.T) {
	t.Parallel()

	session := app.Session{
		ID:     uuid.Must(uuid.NewV4()),
		UserID: user.ID,
	}

	file := bytes.NewBuffer(uuid.Must(uuid.NewV4()).Bytes())

	testCases := []struct {
		name   string
		appErr error
		want   *models.Error
	}{
		{"success", nil, nil},
		{"err_any", errAny, APIError("Internal Server Error")},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, mockApp, client, assert := start(t)
			mockApp.EXPECT().Auth(gomock.Any(), token).Return(&session, nil)

			mockApp.EXPECT().UploadAvatar(gomock.Any(), session, gomock.Any()).Return(tc.appErr)

			params := operations.NewNewAvatarParams().WithUpfile(runtime.NamedReader(tc.name+".txt", file))
			_, err := client.Operations.NewAvatar(params, apiKeyAuth)
			assert.Equal(tc.want, errPayload(err))
		})
	}
}

func TestService_DeleteAvatar(t *testing.T) {
	t.Parallel()

	session := app.Session{
		ID:     uuid.Must(uuid.NewV4()),
		UserID: user.ID,
	}

	fileID := uuid.Must(uuid.NewV4())

	testCases := []struct {
		name   string
		appErr error
		want   *models.Error
	}{
		{"success", nil, nil},
		{"err_any", errAny, APIError("Internal Server Error")},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, mockApp, client, assert := start(t)
			mockApp.EXPECT().Auth(gomock.Any(), token).Return(&session, nil)

			mockApp.EXPECT().DeleteAvatar(gomock.Any(), session, fileID).Return(tc.appErr)

			params := operations.NewDeleteAvatarParams().WithFileID(fileID.String())
			_, err := client.Operations.DeleteAvatar(params, apiKeyAuth)
			assert.Equal(tc.want, errPayload(err))
		})
	}
}
