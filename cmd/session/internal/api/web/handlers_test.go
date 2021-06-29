package web_test

import (
	"net"
	"testing"
	"time"

	web2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/api/web"
	operations2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/api/web/generated/client/operations"
	models2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/api/web/generated/models"
	app2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/app"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
)

func TestService_Login(t *testing.T) {
	t.Parallel()

	var (
		token = app2.Token{
			Value: "token",
		}
		user = app2.User{
			ID:    uuid.Must(uuid.NewV4()),
			Email: "email@email.com",
			Name:  "password",
		}
	)

	testCases := map[string]struct {
		email, pass string
		user        *app2.User
		token       *app2.Token
		appErr      error
		want        *models2.User
		wantErr     *models2.Error
	}{
		"success": {
			user.Email, "password",
			&user, &token, nil, web2.User(&user), nil,
		},
		"err_not_found": {
			"notExist@email.com", "password",
			nil, nil, app2.ErrNotFound, nil, APIError(app2.ErrNotFound.Error()),
		},
		"err_not_valid_password": {
			user.Email, "notValidPass",
			nil, nil, app2.ErrNotValidPassword, nil, APIError(app2.ErrNotValidPassword.Error()),
		},
		"err_any": {
			"randomEmail@email.com", "notValidPass",
			nil, nil, errAny, nil, APIError("Internal Server Error"),
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, mockApp, client, assert := start(t)

			mockApp.EXPECT().Login(gomock.Any(), tc.email, tc.pass, gomock.Any()).Return(tc.user, tc.token, tc.appErr)

			email := models2.Email(tc.email)
			password := models2.Password(tc.pass)

			params := operations2.NewLoginParams().
				WithArgs(&models2.LoginParam{
					Email:    &email,
					Password: &password,
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

	session := app2.Session{
		ID: "id",
		Origin: app2.Origin{
			IP:        net.ParseIP("::1"),
			UserAgent: "Go-http-client/1.1",
		},
		Token: app2.Token{
			Value: "token",
		},
		UserID:    user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	testCases := []struct {
		name   string
		appErr error
		want   *models2.Error
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
			mockApp.EXPECT().Session(gomock.Any(), token).Return(&session, nil)

			params := operations2.NewLogoutParams()
			_, err := client.Operations.Logout(params, apiKeyAuth)
			assert.Equal(tc.want, errPayload(err))
		})
	}
}
