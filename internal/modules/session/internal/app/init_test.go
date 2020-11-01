package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Meat-Hook/back-template/internal/modules/session/internal/app"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

var (
	ctx    = context.Background()
	errAny = errors.New("any error")
)

type mocks struct {
	users *MockUsers
	repo  *MockRepo
	id    *MockID
	auth  *MockAuth
}

func start(t *testing.T) (*app.Module, *mocks, *require.Assertions) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockRepo := NewMockRepo(ctrl)
	mockUsers := NewMockUsers(ctrl)
	mockID := NewMockID(ctrl)
	mockAuth := NewMockAuth(ctrl)

	module := app.New(mockRepo, mockUsers, mockAuth, mockID)

	mocks := &mocks{
		users: mockUsers,
		repo:  mockRepo,
		id:    mockID,
		auth:  mockAuth,
	}

	return module, mocks, require.New(t)
}
