package session_test

import (
	"context"
	"errors"
	"testing"

	client2 "github.com/Meat-Hook/back-template/internal/cmd/session/client"
	app2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/app"
	"github.com/gofrs/uuid"
)

var (
	ctx    = context.Background()
	errAny = errors.New("any err")
)

func TestClient_Session(t *testing.T) {
	t.Parallel()

	sessionInfo := &app2.Session{
		ID:     "id",
		UserID: uuid.Must(uuid.NewV4()),
	}

	testCases := map[string]struct {
		token   string
		want    *app2.Session
		wantErr error
	}{
		"success":       {"validToken", sessionInfo, nil},
		"err_not_found": {"notFoundToken", nil, app2.ErrNotFound},
		"err_any":       {"notValidToken", nil, errAny},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc, mock, assert := start(t)
			mock.EXPECT().Session(ctx, tc.token).Return(convert(tc.want), tc.wantErr)

			res, err := svc.Session(ctx, tc.token)
			assert.Equal(tc.want, res)
			assert.ErrorIs(err, tc.wantErr)
		})
	}
}

func convert(want *app2.Session) *client2.Session {
	if want == nil {
		return nil
	}

	return &client2.Session{
		ID:     want.ID,
		UserID: want.UserID,
	}
}
