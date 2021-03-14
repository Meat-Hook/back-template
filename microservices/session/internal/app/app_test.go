package app_test

import (
	"net"
	"testing"

	"github.com/Meat-Hook/back-template/microservices/session/internal/app"
	"github.com/gofrs/uuid"
)

func TestModule_Login(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	const (
		email = `email@mail.com`
		pass  = `pass`
		id    = `sessionID`

		notValidEmail = `notExist@email.com`

		id2                 = `sessionID2`
		errSaveSessionEmail = `errSaveSession@email.com`
	)

	var (
		origin = app.Origin{
			IP:        net.ParseIP("192.100.10.4"),
			UserAgent: "UserAgent",
		}
		user = app.User{
			ID:    uuid.Must(uuid.NewV4()),
			Email: email,
			Name:  "username",
		}
		user2 = app.User{
			ID:    uuid.Must(uuid.NewV4()),
			Email: errSaveSessionEmail,
			Name:  "username",
		}
		token = app.Token{
			Value: "token",
		}
		token2 = app.Token{
			Value: "token2",
		}
		session = app.Session{
			ID:     id,
			Origin: origin,
			Token:  token,
			UserID: user.ID,
		}
		errSaveSession = app.Session{
			ID:     id2,
			Origin: origin,
			Token:  token2,
			UserID: user2.ID,
		}
	)

	mocks.users.EXPECT().Access(ctx, email, pass).Return(&user, nil)
	mocks.users.EXPECT().Access(ctx, errSaveSessionEmail, pass).Return(&user2, nil)
	mocks.users.EXPECT().Access(ctx, notValidEmail, pass).Return(nil, app.ErrNotFound)
	mocks.id.EXPECT().New().Return(id)
	mocks.id.EXPECT().New().Return(id2)
	mocks.auth.EXPECT().Token(app.Subject{SessionID: id}).Return(&token, nil)
	mocks.auth.EXPECT().Token(app.Subject{SessionID: id2}).Return(&token2, nil)
	mocks.repo.EXPECT().Save(ctx, session).Return(nil)
	mocks.repo.EXPECT().Save(ctx, errSaveSession).Return(errAny)

	testCases := map[string]struct {
		email, password string
		want            *app.User
		wantToken       *app.Token
		wantErr         error
	}{
		"success":       {email, pass, &user, &token, nil},
		"err_any":       {errSaveSessionEmail, pass, nil, nil, errAny},
		"err_not_found": {notValidEmail, pass, nil, nil, app.ErrNotFound},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			resUser, resToken, err := module.Login(ctx, tc.email, tc.password, origin)
			assert.Equal(tc.want, resUser)
			assert.Equal(tc.wantToken, resToken)
			assert.ErrorIs(err, tc.wantErr)
		})
	}
}

func TestModule_Logout(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	session := app.Session{
		ID: "id",
		Origin: app.Origin{
			IP:        net.ParseIP("192.100.10.4"),
			UserAgent: "UserAgent",
		},
		Token: app.Token{
			Value: "token",
		},
		UserID: uuid.Must(uuid.NewV4()),
	}

	mocks.repo.EXPECT().Delete(ctx, session.ID).Return(nil)

	testCases := map[string]struct {
		session *app.Session
		want    error
	}{
		"success": {&session, nil},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			err := module.Logout(ctx, *tc.session)
			assert.Equal(tc.want, err)
		})
	}
}

func TestModule_Session(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	var (
		token          = "token"
		successSubject = app.Subject{SessionID: "ID"}
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
		subjectForNotFoundToken = app.Subject{SessionID: "NOT_FOUND"}

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
