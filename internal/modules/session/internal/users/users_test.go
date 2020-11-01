package users_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Meat-Hook/back-template/internal/modules/session/internal/app"
	user "github.com/Meat-Hook/back-template/internal/modules/user/client"
)

var (
	ctx    = context.Background()
	errAny = errors.New("any err")
)

func TestClient_Access(t *testing.T) {
	t.Parallel()

	svc, mock, assert := start(t)

	userInfo := &app.User{
		ID:    1,
		Email: "email@mail.com",
		Name:  "username",
	}

	testCases := map[string]struct {
		email, pass string
		want        *app.User
		wantErr     error
	}{
		"success":            {userInfo.Email, "pass", userInfo, nil},
		"err_not_found":      {"notFound@email.com", "pass", nil, app.ErrNotFound},
		"err_not_valid_pass": {userInfo.Email, "notValidPass", nil, app.ErrNotValidPassword},
		"err_any":            {"emailNotValid", "", nil, errAny},
	}

	mock.EXPECT().Access(ctx, userInfo.Email, "pass").Return(&user.User{
		ID:    1,
		Email: userInfo.Email,
		Name:  userInfo.Name,
	}, nil)

	mock.EXPECT().Access(ctx, "notFound@email.com", "pass").Return(nil, user.ErrNotFound)
	mock.EXPECT().Access(ctx, userInfo.Email, "notValidPass").Return(nil, user.ErrNotValidPass)
	mock.EXPECT().Access(ctx, "emailNotValid", "").Return(nil, errAny)

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {

			res, err := svc.Access(ctx, tc.email, tc.pass)
			assert.Equal(tc.want, res)
			assert.Equal(tc.wantErr, errors.Unwrap(err))
		})
	}
}
