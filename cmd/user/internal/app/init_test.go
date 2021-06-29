package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Meat-Hook/back-template/cmd/user/internal/app"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

var (
	ctx    = context.Background()
	errAny = errors.New("any error")
)

type mocks struct {
	hasher *MockHasher
	repo   *MockRepo
	auth   *MockAuth
}

func start(t *testing.T) (*app.Module, *mocks, *require.Assertions) {
	t.Helper()
	ctrl := gomock.NewController(t)

	mockRepo := NewMockRepo(ctrl)
	mockHasher := NewMockHasher(ctrl)
	mockAuth := NewMockAuth(ctrl)

	module := app.New(mockRepo, mockHasher, mockAuth)

	mocks := &mocks{
		hasher: mockHasher,
		repo:   mockRepo,
		auth:   mockAuth,
	}

	return module, mocks, require.New(t)
}
