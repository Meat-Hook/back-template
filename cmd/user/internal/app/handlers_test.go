package app_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"github.com/Meat-Hook/back-template/cmd/user/internal/app"
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

	testCases := []struct {
		name  string
		email string
		want  error
	}{
		{"success", free, nil},
		{"err_email_exist", exist, app.ErrEmailExist},
		{"err_any", any, errAny},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := module.VerificationEmail(ctx, tc.email)
			assert.ErrorIs(err, tc.want)
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

	testCases := []struct {
		name     string
		username string
		want     error
	}{
		{"success", free, nil},
		{"err_username_exist", exist, app.ErrUsernameExist},
		{"err_any", any, errAny},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := module.VerificationUsername(ctx, tc.username)
			assert.ErrorIs(err, tc.want)
		})
	}
}

func TestModule_CreateUser(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	var (
		pass          = `pass`
		unknownPass   = `unknownPass`
		email         = `email`
		notValidEmail = `emailNotValid`
		username      = `username`
		existUserName = `existUsername`
		wantID        = uuid.Must(uuid.NewV4())
	)

	mocks.hasher.EXPECT().Hashing(pass).Return([]byte(pass), nil).Times(2)
	mocks.repo.EXPECT().Save(ctx, app.User{
		Email:    email,
		Name:     username,
		PassHash: []byte(pass),
	}).Return(wantID, nil)

	mocks.repo.EXPECT().Save(ctx, app.User{
		Email:    email,
		Name:     existUserName,
		PassHash: []byte(pass),
	}).Return(uuid.Nil, app.ErrUsernameExist)
	mocks.hasher.EXPECT().Hashing(unknownPass).Return(nil, errAny)

	testCases := []struct {
		name     string
		email    string
		username string
		password string
		want     uuid.UUID
		wantErr  error
	}{
		{"success", email, username, pass, wantID, nil},
		{"err_save_user", email, existUserName, pass, uuid.Nil, app.ErrUsernameExist},
		{"err_hashing", notValidEmail, username, unknownPass, uuid.Nil, errAny},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res, err := module.CreateUser(ctx, tc.email, tc.username, tc.password)
			assert.Equal(tc.want, res)
			assert.ErrorIs(err, tc.wantErr)
		})
	}
}

func TestModule_UserByID(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	user := &app.User{
		ID:        uuid.Must(uuid.NewV4()),
		Email:     "email@mail.com",
		Name:      "username",
		PassHash:  []byte{12, 12, 34, 124, 19},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mocks.repo.EXPECT().ByID(ctx, user.ID).Return(user, nil)

	testCases := []struct {
		name    string
		userID  uuid.UUID
		want    *app.User
		wantErr error
	}{
		{"success", user.ID, user, nil},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res, err := module.UserByID(ctx, app.Session{}, tc.userID)
			assert.Equal(tc.want, res)
			assert.ErrorIs(err, tc.wantErr)
		})
	}
}

func TestModule_DeleteUser(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	session := &app.Session{
		ID:     uuid.Must(uuid.NewV4()),
		UserID: uuid.Must(uuid.NewV4()),
	}

	mocks.repo.EXPECT().Delete(ctx, session.UserID).Return(nil)

	testCases := []struct {
		name    string
		session *app.Session
		want    error
	}{
		{"success", session, nil},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := module.DeleteUser(ctx, *tc.session)
			assert.ErrorIs(err, tc.want)
		})
	}
}

func TestModule_UpdateUsername(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	session := &app.Session{
		ID:     uuid.Must(uuid.NewV4()),
		UserID: uuid.Must(uuid.NewV4()),
	}
	notValidSession := &app.Session{
		ID:     uuid.Must(uuid.NewV4()),
		UserID: uuid.Must(uuid.NewV4()),
	}
	user := &app.User{
		ID:        uuid.Must(uuid.NewV4()),
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

	testCases := []struct {
		name     string
		session  *app.Session
		username string
		want     error
	}{
		{"success", session, newUsername, nil},
		{"err_different_username", session, user.Name, app.ErrNotDifferent},
		{"err__not_found", notValidSession, newUsername, app.ErrNotFound},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := module.UpdateUsername(ctx, *tc.session, tc.username)
			assert.ErrorIs(err, tc.want)
		})
	}
}

func TestModule_UpdatePassword(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	session := &app.Session{
		ID:     uuid.Must(uuid.NewV4()),
		UserID: uuid.Must(uuid.NewV4()),
	}
	notValidSession := &app.Session{
		ID:     uuid.Must(uuid.NewV4()),
		UserID: uuid.Must(uuid.NewV4()),
	}
	user := &app.User{
		ID:        uuid.Must(uuid.NewV4()),
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

	testCases := []struct {
		name             string
		session          *app.Session
		oldPass, newPass string
		want             error
	}{
		{"success", session, string(user.PassHash), newPass, nil},
		{"err_hashing", session, string(user.PassHash), notValidPass, errAny},
		{"err_different_pass", session, string(user.PassHash), string(user.PassHash), app.ErrNotDifferent},
		{"err_not_valid_pass", session, notValidPass, newPass, app.ErrNotValidPassword},
		{"err_not_found", notValidSession, "", "", app.ErrNotFound},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := module.UpdatePassword(ctx, *tc.session, tc.oldPass, tc.newPass)
			assert.ErrorIs(err, tc.want)
		})
	}
}

func TestModule_ListUserByUsername(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	user := app.User{
		ID:        uuid.Must(uuid.NewV4()),
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

	testCases := []struct {
		name      string
		username  string
		want      []app.User
		wantTotal int
		wantErr   error
	}{
		{"success", user.Name, []app.User{user}, 1, nil},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res, total, err := module.ListUserByUsername(ctx, app.Session{}, tc.username, p)
			assert.ErrorIs(err, tc.wantErr)
			assert.Equal(tc.want, res)
			assert.Equal(tc.wantTotal, total)
		})
	}
}

func TestModule_Auth(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	session := &app.Session{
		ID:     uuid.Must(uuid.NewV4()),
		UserID: uuid.Must(uuid.NewV4()),
	}

	const token = "token"

	mocks.auth.EXPECT().Session(ctx, token).Return(session, nil)

	testCases := []struct {
		name    string
		token   string
		want    *app.Session
		wantErr error
	}{
		{"success", token, session, nil},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res, err := module.Auth(ctx, tc.token)
			assert.ErrorIs(err, tc.wantErr)
			assert.Equal(tc.want, res)
		})
	}
}

func TestModule_Login(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	user := &app.User{
		ID:        uuid.Must(uuid.NewV4()),
		Email:     "email@mail.com",
		Name:      "username",
		PassHash:  []byte("pass"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	token := &app.Token{
		Value: "auth-token",
	}

	const (
		notValidPass = `notValidPass`
		unknownEmail = `email`
	)

	mocks.auth.EXPECT().NewSession(ctx, user.ID, origin).Return(token, nil)
	mocks.repo.EXPECT().ByEmail(ctx, user.Email).Return(user, nil).Times(2)
	mocks.repo.EXPECT().ByEmail(ctx, unknownEmail).Return(nil, app.ErrNotFound)
	mocks.hasher.EXPECT().Compare(user.PassHash, user.PassHash).Return(true)
	mocks.hasher.EXPECT().Compare(user.PassHash, []byte(notValidPass)).Return(false)

	testCases := []struct {
		name    string
		email   string
		pass    string
		want    *app.Token
		wantErr error
	}{
		{"success", user.Email, string(user.PassHash), token, nil},
		{"err_not_valid", user.Email, notValidPass, nil, app.ErrNotValidPassword},
		{"err_not_found", unknownEmail, "", nil, app.ErrNotFound},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res, err := module.Login(ctx, tc.email, tc.pass, origin)
			assert.ErrorIs(err, tc.wantErr)
			assert.Equal(tc.want, res)
		})
	}
}

func TestModule_Logout(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	session := &app.Session{
		ID:     uuid.Must(uuid.NewV4()),
		UserID: uuid.Must(uuid.NewV4()),
	}

	mocks.auth.EXPECT().RemoveSession(ctx, session.ID).Return(nil)

	testCases := []struct {
		name    string
		session *app.Session
		wantErr error
	}{
		{"success", session, nil},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := module.Logout(ctx, *tc.session)
			assert.ErrorIs(err, tc.wantErr)
		})
	}
}

func TestModule_UploadAvatar(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	fileID := uuid.Must(uuid.NewV4())
	userNotFoundID := uuid.Must(uuid.NewV4())

	userWithoutAvatar := app.User{
		ID:        uuid.Must(uuid.NewV4()),
		Email:     "email@mail.com",
		Name:      "username",
		PassHash:  []byte{12, 12, 34, 124, 19},
		Avatars:   []uuid.UUID{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	userWithAvatar := app.User{
		ID:        userWithoutAvatar.ID,
		Email:     "email@mail.com",
		Name:      "username",
		PassHash:  []byte{12, 12, 34, 124, 19},
		Avatars:   []uuid.UUID{fileID},
		CreatedAt: userWithoutAvatar.CreatedAt,
		UpdatedAt: userWithoutAvatar.UpdatedAt,
	}

	correctFile := bytes.NewBuffer(uuid.Must(uuid.NewV4()).Bytes())

	session := &app.Session{
		ID:     uuid.Must(uuid.NewV4()),
		UserID: userWithoutAvatar.ID,
	}

	mocks.repo.EXPECT().ByID(ctx, userWithoutAvatar.ID).Return(&userWithoutAvatar, nil).Times(2)
	mocks.repo.EXPECT().ByID(ctx, userNotFoundID).Return(nil, app.ErrNotFound)
	mocks.file.EXPECT().Upload(ctx, correctFile).Return(fileID, nil)
	mocks.file.EXPECT().Upload(ctx, nil).Return(uuid.Nil, errAny)
	mocks.repo.EXPECT().Update(ctx, userWithAvatar).Return(nil)

	testCases := []struct {
		name    string
		session *app.Session
		file    io.Reader
		wantErr error
	}{
		{"success", session, correctFile, nil},
		{"err_upload_file", session, nil, errAny},
		{"err_user_not_found", &app.Session{UserID: userNotFoundID}, nil, app.ErrNotFound},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := module.UploadAvatar(ctx, *tc.session, tc.file)
			assert.ErrorIs(err, tc.wantErr)
		})
	}
}

func TestModule_DeleteAvatar(t *testing.T) {
	t.Parallel()

	module, mocks, assert := start(t)

	fileID := uuid.Must(uuid.NewV4())
	fileID2 := uuid.Must(uuid.NewV4())
	fileID3 := uuid.Must(uuid.NewV4())

	userNotFoundID := uuid.Must(uuid.NewV4())

	user := app.User{
		ID:        uuid.Must(uuid.NewV4()),
		Email:     "email@mail.com",
		Name:      "username",
		PassHash:  []byte{12, 12, 34, 124, 19},
		Avatars:   []uuid.UUID{fileID, fileID2, fileID3},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	userWithoutSecondAvatar := app.User{
		ID:        user.ID,
		Email:     "email@mail.com",
		Name:      "username",
		PassHash:  []byte{12, 12, 34, 124, 19},
		Avatars:   []uuid.UUID{fileID, fileID3},
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	session := &app.Session{
		ID:     uuid.Must(uuid.NewV4()),
		UserID: user.ID,
	}

	mocks.repo.EXPECT().ByID(ctx, user.ID).Return(&user, nil).Times(3)
	mocks.repo.EXPECT().ByID(ctx, userNotFoundID).Return(nil, app.ErrNotFound)
	mocks.file.EXPECT().Delete(ctx, fileID2).Return(nil)
	mocks.file.EXPECT().Delete(ctx, fileID3).Return(errAny)
	mocks.repo.EXPECT().Update(ctx, userWithoutSecondAvatar).Return(nil)

	testCases := []struct {
		name    string
		session *app.Session
		fileID  uuid.UUID
		wantErr error
	}{
		{"success", session, fileID2, nil},
		{"err_file_not_delete", session, fileID3, errAny},
		{"err_file_not_found", session, uuid.Nil, app.ErrNotFound},
		{"err_user_not_found", &app.Session{UserID: userNotFoundID}, uuid.Nil, app.ErrNotFound},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := module.DeleteAvatar(ctx, *tc.session, tc.fileID)
			assert.ErrorIs(err, tc.wantErr)
		})
	}
}
