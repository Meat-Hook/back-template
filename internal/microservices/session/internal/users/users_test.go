package users_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Meat-Hook/back-template/internal/microservices/session/internal/app"
	"github.com/Meat-Hook/back-template/internal/microservices/user/client"
)

var (
	ctx    = context.Background()
	errAny = errors.New("any err")
)

func TestClient_Access(t *testing.T) {
	t.Parallel()

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

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc, mock, assert := start(t)

			mock.EXPECT().Access(ctx, tc.email, tc.pass).
				Return(convert(tc.want), tc.wantErr)

			res, err := svc.Access(ctx, tc.email, tc.pass)
			assert.Equal(tc.want, res)
			assert.ErrorIs(err, tc.wantErr)
		})
	}
}

func convert(want *app.User) *client.User {
	if want == nil {
		return nil
	}

	return &client.User{
		ID:    want.ID,
		Email: want.Email,
		Name:  want.Name,
	}
}
