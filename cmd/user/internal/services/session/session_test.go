package session_test

import (
	"testing"

	"github.com/gofrs/uuid"

	"github.com/Meat-Hook/back-template/cmd/session/client"
	"github.com/Meat-Hook/back-template/cmd/user/internal/app"
)

func TestClient_Session(t *testing.T) {
	t.Parallel()

	sessionInfo := &app.Session{
		ID:     uuid.Must(uuid.NewV4()),
		UserID: uuid.Must(uuid.NewV4()),
	}

	testCases := []struct {
		name    string
		token   string
		want    *app.Session
		wantErr error
	}{
		{"success", "validToken", sessionInfo, nil},
		{"err_not_found", "notFoundToken", nil, app.ErrNotFound},
		{"err_any", "notValidToken", nil, errAny},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc, mock, assert := start(t)

			var wantReturn *client.Session
			if tc.want != nil {
				wantReturn = &client.Session{
					ID:     tc.want.ID,
					UserID: tc.want.UserID,
				}
			}
			mock.EXPECT().Session(ctx, tc.token).Return(wantReturn, tc.wantErr)

			res, err := svc.Session(ctx, tc.token)
			assert.Equal(tc.want, res)
			assert.ErrorIs(err, tc.wantErr)
		})
	}
}

func TestClient_NewSession(t *testing.T) {
	t.Parallel()

	var (
		token = &app.Token{
			Value: "token",
		}

		userID = uuid.Must(uuid.NewV4())
	)

	testCases := []struct {
		name    string
		want    *app.Token
		wantErr error
	}{
		{"success", token, nil},
		{"err_any", nil, errAny},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc, mock, assert := start(t)
			var wantReturn *client.Token
			if tc.want != nil {
				wantReturn = &client.Token{
					Value: tc.want.Value,
				}
			}

			mock.EXPECT().NewSession(ctx, userID, origin.IP, origin.UserAgent).Return(wantReturn, tc.wantErr)

			res, err := svc.NewSession(ctx, userID, origin)
			assert.Equal(tc.want, res)
			assert.ErrorIs(err, tc.wantErr)
		})
	}
}

func TestClient_RemoveSession(t *testing.T) {
	t.Parallel()

	sessionID := uuid.Must(uuid.NewV4())

	testCases := []struct {
		name string
		want error
	}{
		{"success", nil},
		{"err_any", errAny},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc, mock, assert := start(t)

			mock.EXPECT().RemoveSession(ctx, sessionID).Return(tc.want)

			err := svc.RemoveSession(ctx, sessionID)
			assert.ErrorIs(err, tc.want)
		})
	}
}
