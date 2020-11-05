// Package app contains all logic of the microservice.
package app

import (
	"context"
	"errors"
	"strings"
	"time"
)

// Errors.
var (
	ErrEmailExist       = errors.New("email exist")
	ErrUsernameExist    = errors.New("username exist")
	ErrNotFound         = errors.New("not found")
	ErrNotDifferent     = errors.New("the values must be different")
	ErrNotValidPassword = errors.New("not valid password")
)

type (
	// Repo interface for user data repository.
	Repo interface {
		// Save adds to the new user in repository.
		// Errors: ErrEmailExist, ErrUsernameExist, unknown.
		Save(context.Context, User) (id int, err error)
		// Update update user info.
		// Errors: ErrUsernameExist, ErrEmailExist, unknown.
		Update(context.Context, User) error
		// Delete removes user from repository by id.
		// Errors: unknown.
		Delete(context.Context, int) error
		// ByID returning user info by id.
		// Errors: ErrNotFound, unknown.
		ByID(context.Context, int) (*User, error)
		// ByEmail returning user info by email.
		// Errors: ErrNotFound, unknown.
		ByEmail(context.Context, string) (*User, error)
		// ByUsername returning user info by username.
		// Errors: ErrNotFound, unknown.
		ByUsername(context.Context, string) (*User, error)
		// ListUserByUsername returning list user info.
		// Errors: unknown.
		ListUserByUsername(context.Context, string, SearchParams) ([]User, int, error)
	}

	// Hasher module responsible for hashing password.
	Hasher interface {
		// Hashing returns the hashed version of the password.
		// Errors: unknown.
		Hashing(password string) ([]byte, error)
		// Compare compares two passwords for matches.
		Compare(hashedPassword []byte, password []byte) bool
	}

	// Notification module for working with alerts for registered users.
	Notification interface {
		// Send sends a message to the user based on their contact information.
		// Errors: unknown.
		Send(ctx context.Context, contact string, msg Message) error
	}

	// Auth module for get user session by token.
	Auth interface {
		// Session returns user session by his token.
		// Errors: ErrNotFound, unknown.
		Session(ctx context.Context, token string) (*Session, error)
	}

	// Message contains notification info.
	Message struct {
		Kind    MessageKind
		Content string
	}

	// MessageKind selects the type of message to be sent.
	MessageKind uint8

	// SearchParams params for search users.
	SearchParams struct {
		Limit  uint
		Offset uint
	}

	// Session contains user session information.
	Session struct {
		ID     string
		UserID int
	}

	// User contains user information.
	User struct {
		ID       int
		Email    string
		Name     string
		PassHash []byte

		CreatedAt time.Time
		UpdatedAt time.Time
	}

	// Module contains business logic for user methods.
	Module struct {
		user         Repo
		hash         Hasher
		notification Notification
		auth         Auth
	}
)

// New build and returns new Module for working with user info.
func New(r Repo, h Hasher, n Notification, a Auth) *Module {
	return &Module{
		user:         r,
		hash:         h,
		notification: n,
		auth:         a,
	}
}

// Message enums.
const (
	Welcome MessageKind = iota + 1
)

// Message text.
const (
	WelcomeText = `Welcome`
)

// VerificationEmail check exists or not user email.
func (m *Module) VerificationEmail(ctx context.Context, email string) error {
	_, err := m.user.ByEmail(ctx, email)
	switch {
	case errors.Is(err, ErrNotFound):
		return nil
	case err == nil:
		return ErrEmailExist
	default:
		return err
	}
}

// VerificationUsername check exists or not username.
func (m *Module) VerificationUsername(ctx context.Context, username string) error {
	_, err := m.user.ByUsername(ctx, username)
	switch {
	case errors.Is(err, ErrNotFound):
		return nil
	case err == nil:
		return ErrUsernameExist
	default:
		return err
	}
}

// CreateUser create new user by params.
func (m *Module) CreateUser(ctx context.Context, email, username, password string) (int, error) {
	passHash, err := m.hash.Hashing(password)
	if err != nil {
		return 0, err
	}
	email = strings.ToLower(email)

	msg := Message{
		Kind:    Welcome,
		Content: WelcomeText,
	}

	err = m.notification.Send(ctx, email, msg)
	if err != nil {
		return 0, err
	}

	newUser := User{
		Email:    email,
		Name:     username,
		PassHash: passHash,
	}

	userID, err := m.user.Save(ctx, newUser)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

// UserByID get user by id.
func (m *Module) UserByID(ctx context.Context, _ Session, userID int) (*User, error) {
	return m.user.ByID(ctx, userID)
}

// DeleteUser remove user from repo.
func (m *Module) DeleteUser(ctx context.Context, session Session) error {
	return m.user.Delete(ctx, session.UserID)
}

// UpdateUsername update username.
func (m *Module) UpdateUsername(ctx context.Context, session Session, username string) error {
	user, err := m.user.ByID(ctx, session.UserID)
	if err != nil {
		return err
	}

	if user.Name == username {
		return ErrNotDifferent
	}
	user.Name = username

	return m.user.Update(ctx, *user)
}

// UpdatePassword update user password.
func (m *Module) UpdatePassword(ctx context.Context, session Session, oldPass, newPass string) error {
	user, err := m.user.ByID(ctx, session.UserID)
	if err != nil {
		return err
	}

	if !m.hash.Compare(user.PassHash, []byte(oldPass)) {
		return ErrNotValidPassword
	}

	if m.hash.Compare(user.PassHash, []byte(newPass)) {
		return ErrNotDifferent
	}

	passHash, err := m.hash.Hashing(newPass)
	if err != nil {
		return err
	}
	user.PassHash = passHash

	return m.user.Update(ctx, *user)
}

// ListUserByUsername get users by username.
func (m *Module) ListUserByUsername(ctx context.Context, _ Session, username string, p SearchParams) ([]User, int, error) {
	return m.user.ListUserByUsername(ctx, username, p)
}

// Auth get user session by token.
func (m *Module) Auth(ctx context.Context, token string) (*Session, error) {
	return m.auth.Session(ctx, token)
}

// Access finds a user by email and compares his password to allow access.
func (m *Module) Access(ctx context.Context, email, password string) (*User, error) {
	email = strings.ToLower(email)
	user, err := m.user.ByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if !m.hash.Compare(user.PassHash, []byte(password)) {
		return nil, ErrNotValidPassword
	}

	return user, nil
}
