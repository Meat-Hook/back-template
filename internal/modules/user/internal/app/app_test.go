package app_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Meat-Hook/back-template/internal/modules/user/internal/app"
)

func TestModule_VerificationEmail(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	const (
		exist = "exist@mail.com"
		free  = "free@mail.com"
		any   = "any@mail.com"
	)

	mocks.repo.EXPECT().ByEmail(ctx, exist).Return(&app.User{}, nil)
	mocks.repo.EXPECT().ByEmail(ctx, free).Return(nil, app.ErrNotFound)
	mocks.repo.EXPECT().ByEmail(ctx, any).Return(nil, errAny)

	testCases := map[string]struct {
		email string
		want  error
	}{
		"success":   {free, nil},
		"exist":     {exist, app.ErrEmailExist},
		"any_error": {any, errAny},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			err := module.VerificationEmail(ctx, tc.email)
			assert.Equal(tc.want, err)
		})
	}
}

func TestModule_VerificationUsername(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	const (
		exist = "exist"
		free  = "free"
		any   = "any"
	)

	mocks.repo.EXPECT().ByUsername(ctx, exist).Return(&app.User{}, nil)
	mocks.repo.EXPECT().ByUsername(ctx, free).Return(nil, app.ErrNotFound)
	mocks.repo.EXPECT().ByUsername(ctx, any).Return(nil, errAny)

	testCases := map[string]struct {
		username string
		want     error
	}{
		"success":   {free, nil},
		"exist":     {exist, app.ErrUsernameExist},
		"any_error": {any, errAny},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			err := module.VerificationUsername(ctx, tc.username)
			assert.Equal(tc.want, err)
		})
	}
}

func TestModule_CreateUser(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	const (
		pass          = `pass`
		unknownPass   = `unknownPass`
		email         = `email`
		notValidEmail = `emailNotValid`
		username      = `username`
		existUserName = `existUsername`
		wantID        = 1
	)

	mocks.hasher.EXPECT().Hashing(pass).Return([]byte(pass), nil).Times(3)
	mocks.notification.EXPECT().Send(ctx, email, app.Message{
		Kind:    app.Welcome,
		Content: app.WelcomeText,
	}).Return(nil).Times(2)
	mocks.repo.EXPECT().Save(ctx, app.User{
		Email:    email,
		Name:     username,
		PassHash: []byte(pass),
	}).Return(wantID, nil)

	mocks.repo.EXPECT().Save(ctx, app.User{
		Email:    email,
		Name:     existUserName,
		PassHash: []byte(pass),
	}).Return(0, app.ErrUsernameExist)

	mocks.notification.EXPECT().Send(ctx, strings.ToLower(notValidEmail), app.Message{
		Kind:    app.Welcome,
		Content: app.WelcomeText,
	}).Return(errAny)

	mocks.hasher.EXPECT().Hashing(unknownPass).Return(nil, errAny)

	testCases := map[string]struct {
		email    string
		username string
		password string
		want     int
		wantErr  error
	}{
		"success":          {email, username, pass, wantID, nil},
		"err_save_user":    {email, existUserName, pass, 0, app.ErrUsernameExist},
		"err_notification": {notValidEmail, username, pass, 0, errAny},
		"err_hashing":      {notValidEmail, username, unknownPass, 0, errAny},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			res, err := module.CreateUser(ctx, tc.email, tc.username, tc.password)
			assert.Equal(tc.want, res)
			assert.Equal(tc.wantErr, err)
		})
	}
}

func TestModule_UserByID(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	user := &app.User{
		ID:        1,
		Email:     "email@mail.com",
		Name:      "username",
		PassHash:  []byte{12, 12, 34, 124, 19},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mocks.repo.EXPECT().ByID(ctx, user.ID).Return(user, nil)

	testCases := map[string]struct {
		userID  int
		want    *app.User
		wantErr error
	}{
		"success": {user.ID, user, nil},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			res, err := module.UserByID(ctx, app.Session{}, tc.userID)
			assert.Equal(tc.wantErr, err)
			assert.Equal(tc.want, res)
		})
	}
}

func TestModule_DeleteUser(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	session := &app.Session{
		ID:     "id",
		UserID: 1,
	}

	mocks.repo.EXPECT().Delete(ctx, session.UserID).Return(nil)

	testCases := map[string]struct {
		session *app.Session
		want    error
	}{
		"success": {session, nil},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			err := module.DeleteUser(ctx, *tc.session)
			assert.Equal(tc.want, err)
		})
	}
}

func TestModule_UpdateUsername(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	session := &app.Session{
		ID:     "id",
		UserID: 1,
	}
	notValidSession := &app.Session{
		ID:     "id2",
		UserID: 2,
	}
	user := &app.User{
		ID:        1,
		Email:     "email@mail.com",
		Name:      "username",
		PassHash:  []byte{1, 2, 3, 34, 5, 6, 7},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	const newUsername = `newUsername`
	updatedUser := *user
	updatedUser.Name = newUsername

	mocks.repo.EXPECT().ByID(ctx, session.UserID).Return(user, nil).Times(2)
	mocks.repo.EXPECT().Update(ctx, updatedUser).Return(nil).Do(func(_ context.Context, _ app.User) {
		user.Name = "username"
	})
	mocks.repo.EXPECT().ByID(ctx, notValidSession.UserID).Return(nil, app.ErrNotFound)

	testCases := map[string]struct {
		session  *app.Session
		username string
		want     error
	}{
		"success":         {session, newUsername, nil},
		"usernames_equal": {session, user.Name, app.ErrNotDifferent},
		"user_not_found":  {notValidSession, newUsername, app.ErrNotFound},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			err := module.UpdateUsername(ctx, *tc.session, tc.username)
			assert.Equal(tc.want, err)
		})
	}
}

func TestModule_UpdatePassword(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	session := &app.Session{
		ID:     "id",
		UserID: 1,
	}
	notValidSession := &app.Session{
		ID:     "id2",
		UserID: 2,
	}
	user := &app.User{
		ID:        1,
		Email:     "email@mail.com",
		Name:      "username",
		PassHash:  []byte("pass"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	const (
		newPass      = `newPass`
		notValidPass = `notValidPass`
	)

	updatedUser := *user
	updatedUser.PassHash = []byte(newPass)

	mocks.repo.EXPECT().ByID(ctx, session.UserID).Return(user, nil).Times(4)
	mocks.hasher.EXPECT().Compare(user.PassHash, user.PassHash).Return(true).Times(3)
	mocks.hasher.EXPECT().Compare(user.PassHash, []byte(newPass)).Return(false)
	mocks.hasher.EXPECT().Compare(user.PassHash, []byte(notValidPass)).Return(false)
	mocks.hasher.EXPECT().Compare(user.PassHash, user.PassHash).Return(true)
	mocks.hasher.EXPECT().Compare(user.PassHash, []byte(notValidPass)).Return(false)
	mocks.hasher.EXPECT().Hashing(newPass).Return([]byte(newPass), nil)
	mocks.hasher.EXPECT().Hashing(notValidPass).Return(nil, errAny)

	mocks.repo.EXPECT().Update(ctx, updatedUser).Return(nil).Do(func(_ context.Context, _ app.User) {
		user.PassHash = []byte("pass")
	})
	mocks.repo.EXPECT().ByID(ctx, notValidSession.UserID).Return(nil, app.ErrNotFound)

	testCases := map[string]struct {
		session          *app.Session
		oldPass, newPass string
		want             error
	}{
		"success":            {session, string(user.PassHash), newPass, nil},
		"err_hashing":        {session, string(user.PassHash), notValidPass, errAny},
		"err_different_pass": {session, string(user.PassHash), string(user.PassHash), app.ErrNotDifferent},
		"err_not_valid_pass": {session, notValidPass, newPass, app.ErrNotValidPassword},
		"err_not_found":      {notValidSession, "", "", app.ErrNotFound},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			err := module.UpdatePassword(ctx, *tc.session, tc.oldPass, tc.newPass)
			assert.Equal(tc.want, err)
		})
	}
}

func TestModule_ListUserByUsername(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	user := app.User{
		ID:        1,
		Email:     "email@mail.com",
		Name:      "username",
		PassHash:  []byte{12, 12, 34, 124, 19},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	p := app.SearchParams{
		Limit:  5,
		Offset: 0,
	}

	mocks.repo.EXPECT().ListUserByUsername(ctx, user.Name, p).Return([]app.User{user}, 1, nil)

	testCases := map[string]struct {
		username  string
		want      []app.User
		wantTotal int
		wantErr   error
	}{
		"success": {user.Name, []app.User{user}, 1, nil},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			res, total, err := module.ListUserByUsername(ctx, app.Session{}, tc.username, p)
			assert.Equal(tc.wantErr, err)
			assert.Equal(tc.want, res)
			assert.Equal(tc.wantTotal, total)
		})
	}
}

func TestModule_Auth(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	session := &app.Session{
		ID:     "id",
		UserID: 1,
	}

	const token = "token"

	mocks.auth.EXPECT().Session(ctx, token).Return(session, nil)

	testCases := map[string]struct {
		token   string
		want    *app.Session
		wantErr error
	}{
		"success": {token, session, nil},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			res, err := module.Auth(ctx, tc.token)
			assert.Equal(tc.wantErr, err)
			assert.Equal(tc.want, res)
		})
	}
}

func TestModule_Access(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	user := &app.User{
		ID:        1,
		Email:     "email@mail.com",
		Name:      "username",
		PassHash:  []byte("pass"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	const (
		notValidPass = `notValidPass`
		unknownEmail = `email`
	)

	mocks.repo.EXPECT().ByEmail(ctx, user.Email).Return(user, nil).Times(2)
	mocks.repo.EXPECT().ByEmail(ctx, unknownEmail).Return(nil, app.ErrNotFound)
	mocks.hasher.EXPECT().Compare(user.PassHash, user.PassHash).Return(true)
	mocks.hasher.EXPECT().Compare(user.PassHash, []byte(notValidPass)).Return(false)

	testCases := map[string]struct {
		email   string
		pass    string
		want    *app.User
		wantErr error
	}{
		"success":            {user.Email, string(user.PassHash), user, nil},
		"err_pass_not_valid": {user.Email, notValidPass, nil, app.ErrNotValidPassword},
		"err_user_not_found": {unknownEmail, "", nil, app.ErrNotFound},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			res, err := module.Access(ctx, tc.email, tc.pass)
			assert.Equal(tc.wantErr, err)
			assert.Equal(tc.want, res)
		})
	}
}
