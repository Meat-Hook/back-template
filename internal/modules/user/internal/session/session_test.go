package session_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Meat-Hook/back-template/internal/modules/session/client"
	session "github.com/Meat-Hook/back-template/internal/modules/session/client"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/app"
)

var (
	ctx    = context.Background()
	errAny = errors.New("any err")
)

func TestClient_Session(t *testing.T) {
	t.Parallel()

	svc, mock, assert := start(t)

	sessionInfo := &app.Session{
		ID:     "id",
		UserID: 1,
	}

	testCases := map[string]struct {
		token   string
		want    *app.Session
		wantErr error
	}{
		"success":       {"validToken", sessionInfo, nil},
		"err_not_found": {"notFoundToken", nil, app.ErrNotFound},
		"err_any":       {"notValidToken", nil, errAny},
	}

	mock.EXPECT().Session(ctx, "validToken").Return(&client.Session{
		ID:     sessionInfo.ID,
		UserID: sessionInfo.UserID,
	}, nil)
	mock.EXPECT().Session(ctx, "notFoundToken").Return(nil, session.ErrNotFound)
	mock.EXPECT().Session(ctx, "notValidToken").Return(nil, errAny)

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {

			res, err := svc.Session(ctx, tc.token)
			assert.Equal(tc.want, res)
			assert.Equal(tc.wantErr, errors.Unwrap(err))
		})
	}
}
