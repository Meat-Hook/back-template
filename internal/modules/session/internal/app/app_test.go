package app_test

import (
	"net"
	"testing"

	"github.com/Meat-Hook/back-template/internal/modules/session/internal/app"
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
			ID:    1,
			Email: email,
			Name:  "username",
		}
		user2 = app.User{
			ID:    2,
			Email: errSaveSessionEmail,
			Name:  "username",
		}
		token = app.Token{
			Value: "token",
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
			Token:  token,
			UserID: user2.ID,
		}
	)

	mocks.users.EXPECT().Access(ctx, email, pass).Return(&user, nil)
	mocks.users.EXPECT().Access(ctx, errSaveSessionEmail, pass).Return(&user2, nil)
	mocks.users.EXPECT().Access(ctx, notValidEmail, pass).Return(nil, app.ErrNotFound)
	mocks.id.EXPECT().New().Return(id)
	mocks.id.EXPECT().New().Return(id2)
	mocks.auth.EXPECT().Token(app.Subject{SessionID: id}).Return(&token, nil)
	mocks.auth.EXPECT().Token(app.Subject{SessionID: id2}).Return(&token, nil)
	mocks.repo.EXPECT().Save(ctx, session).Return(nil)
	mocks.repo.EXPECT().Save(ctx, errSaveSession).Return(errAny)

	testCases := map[string]struct {
		email, password string
		want            *app.User
		wantToken       *app.Token
		wantErr         error
	}{
		"success":          {email, pass, &user, &token, nil},
		"err_save_session": {errSaveSessionEmail, pass, nil, nil, errAny},
		"err_user_access":  {notValidEmail, pass, nil, nil, app.ErrNotFound},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			resUser, resToken, err := module.Login(ctx, tc.email, tc.password, origin)
			assert.Equal(tc.want, resUser)
			assert.Equal(tc.wantToken, resToken)
			assert.Equal(tc.wantErr, err)
		})
	}
}

func TestModule_Logout(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	var (
		session = app.Session{
			ID: "id",
			Origin: app.Origin{
				IP:        net.ParseIP("192.100.10.4"),
				UserAgent: "UserAgent",
			},
			Token: app.Token{
				Value: "token",
			},
			UserID: 1,
		}
	)

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
			UserID: 1,
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
		"success":             {token, &session, nil},
		"err_session_by_id":   {tokenNotFound, nil, app.ErrNotFound},
		"err_not_valid_token": {notValidToken, nil, app.ErrInvalidToken},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			res, err := module.Session(ctx, tc.token)
			assert.Equal(tc.want, res)
			assert.Equal(tc.wantErr, err)
		})
	}
}
