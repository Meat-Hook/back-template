package app_test

import (
	"net"
	"testing"

	"github.com/gofrs/uuid"

	"github.com/Meat-Hook/back-template/cmd/session/internal/app"
)

func TestModule_NewSession(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	var (
		id  = uuid.Must(uuid.NewV4())
		id2 = uuid.Must(uuid.NewV4())

		origin = app.Origin{
			IP:        net.ParseIP("192.100.10.4"),
			UserAgent: "UserAgent",
		}
		userID1 = uuid.Must(uuid.NewV4())
		userID2 = uuid.Must(uuid.NewV4())
		token   = app.Token{
			Value: "token",
		}
		token2 = app.Token{
			Value: "token2",
		}
		session = app.Session{
			ID:     id,
			Origin: origin,
			Token:  token,
			UserID: userID1,
		}
		errSaveSession = app.Session{
			ID:     id2,
			Origin: origin,
			Token:  token2,
			UserID: userID2,
		}
	)

	mocks.id.EXPECT().New().Return(id)
	mocks.id.EXPECT().New().Return(id2)
	mocks.auth.EXPECT().Token(app.Subject{SessionID: id}).Return(&token, nil)
	mocks.auth.EXPECT().Token(app.Subject{SessionID: id2}).Return(&token2, nil)
	mocks.repo.EXPECT().Save(ctx, session).Return(nil)
	mocks.repo.EXPECT().Save(ctx, errSaveSession).Return(errAny)

	testCases := map[string]struct {
		userID  uuid.UUID
		want    *app.Token
		wantErr error
	}{
		"success":       {userID1, &token, nil},
		"err_any":       {userID2, nil, errAny},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			resToken, err := module.NewSession(ctx, tc.userID, origin)
			assert.Equal(tc.want, resToken)
			assert.ErrorIs(err, tc.wantErr)
		})
	}
}

func TestModule_RemoveSession(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	id := uuid.Must(uuid.NewV4())
	mocks.repo.EXPECT().Delete(ctx, id).Return(nil)

	testCases := map[string]struct {
		session uuid.UUID
		want    error
	}{
		"success": {id, nil},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			err := module.RemoveSession(ctx, id)
			assert.Equal(tc.want, err)
		})
	}
}

func TestModule_Session(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	var (
		token          = "token"
		successSubject = app.Subject{SessionID: uuid.Must(uuid.NewV4())}
		session        = app.Session{
			ID: successSubject.SessionID,
			Origin: app.Origin{
				IP:        net.ParseIP("192.100.10.4"),
				UserAgent: "UserAgent",
			},
			Token: app.Token{
				Value: token,
			},
			UserID: uuid.Must(uuid.NewV4()),
		}

		tokenNotFound           = "tokenNotFound"
		subjectForNotFoundToken = app.Subject{SessionID: uuid.Must(uuid.NewV4())}

		notValidToken = "notValidToken"
	)

	mocks.auth.EXPECT().Subject(token).Return(&successSubject, nil)
	mocks.auth.EXPECT().Subject(tokenNotFound).Return(&subjectForNotFoundToken, nil)
	mocks.auth.EXPECT().Subject(notValidToken).Return(nil, app.ErrInvalidToken)
	mocks.repo.EXPECT().ByID(ctx, successSubject.SessionID).Return(&session, nil)
	mocks.repo.EXPECT().ByID(ctx, subjectForNotFoundToken.SessionID).Return(nil, app.ErrNotFound)

	testCases := map[string]struct {
		token   string
		want    *app.Session
		wantErr error
	}{
		"success":           {token, &session, nil},
		"err_not_found":     {tokenNotFound, nil, app.ErrNotFound},
		"err_invalid_token": {notValidToken, nil, app.ErrInvalidToken},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			res, err := module.Session(ctx, tc.token)
			assert.Equal(tc.want, res)
			assert.ErrorIs(err, tc.wantErr)
		})
	}
}
