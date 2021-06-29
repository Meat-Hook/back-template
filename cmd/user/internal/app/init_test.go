package app_test

import (
	"context"
	"errors"
	"testing"

	app2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/app"
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

func start(t *testing.T) (*app2.Module, *mocks, *require.Assertions) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockRepo := NewMockRepo(ctrl)
	mockHasher := NewMockHasher(ctrl)
	mockAuth := NewMockAuth(ctrl)

	module := app2.New(mockRepo, mockHasher, mockAuth)

	mocks := &mocks{
		hasher: mockHasher,
		repo:   mockRepo,
		auth:   mockAuth,
	}

	return module, mocks, require.New(t)
}
