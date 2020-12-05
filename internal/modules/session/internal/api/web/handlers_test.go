package web_test

import (
	"net"
	"testing"
	"time"

	"github.com/Meat-Hook/back-template/internal/modules/session/internal/api/web"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/api/web/generated/client/operations"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/api/web/generated/models"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/app"
	"github.com/golang/mock/gomock"
)

func TestService_Login(t *testing.T) {
	t.Parallel()

	_, mockApp, client, assert := start(t)

	var (
		token = app.Token{
			Value: "token",
		}
		user = app.User{
			ID:    1,
			Email: "email@email.com",
			Name:  "password",
		}
	)

	testCases := map[string]struct {
		email, pass string
		user        *app.User
		token       *app.Token
		appErr      error
		want        *models.User
		wantErr     *models.Error
	}{
		"success": {user.Email, "password",
			&user, &token, nil, web.User(&user), nil},
		"err_not_found": {"notExist@email.com", "password",
			nil, nil, app.ErrNotFound, nil, APIError(app.ErrNotFound.Error())},
		"err_not_valid_password": {user.Email, "notValidPass",
			nil, nil, app.ErrNotValidPassword, nil, APIError(app.ErrNotValidPassword.Error())},
		"err_internal": {"randomEmail@email.com", "notValidPass",
			nil, nil, errAny, nil, APIError("Internal Server Error")},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			mockApp.EXPECT().Login(gomock.Any(), tc.email, tc.pass, gomock.Any()).Return(tc.user, tc.token, tc.appErr)

			params := operations.NewLoginParams().
				WithArgs(&models.LoginParam{
					Email:    models.Email(tc.email),
					Password: models.Password(tc.pass),
				})
			res, err := client.Operations.Login(params)
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

func TestService_Logout(t *testing.T) {
	t.Parallel()

	_, mockApp, client, assert := start(t)

	session := app.Session{
		ID: "id",
		Origin: app.Origin{
			IP:        net.ParseIP("::1"),
			UserAgent: "Go-http-client/1.1",
		},
		Token: app.Token{
			Value: "token",
		},
		UserID:    user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	testCases := []struct {
		name   string
		appErr error
		want   *models.Error
	}{
		{"success", nil, nil},
		{"err_internal", errAny, APIError("Internal Server Error")},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mockApp.EXPECT().Logout(gomock.Any(), session).Return(tc.appErr)
			mockApp.EXPECT().Session(gomock.Any(), token).Return(&session, nil)

			params := operations.NewLogoutParams()
			_, err := client.Operations.Logout(params, apiKeyAuth)
			assert.Equal(tc.want, errPayload(err))
		})
	}
}
