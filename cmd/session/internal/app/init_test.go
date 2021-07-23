package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/Meat-Hook/back-template/cmd/session/internal/app"
)

var (
	ctx    = context.Background()
	errAny = errors.New("any error")
)

type mocks struct {
	repo *MockRepo
	id   *MockID
	auth *MockAuth
}

func start(t *testing.T) (*app.Module, *mocks, *require.Assertions) {
	t.Helper()
	ctrl := gomock.NewController(t)

	mockRepo := NewMockRepo(ctrl)
	mockID := NewMockID(ctrl)
	mockAuth := NewMockAuth(ctrl)

	module := app.New(mockRepo, mockAuth, mockID)

	mocks := &mocks{
		repo: mockRepo,
		id:   mockID,
		auth: mockAuth,
	}

	return module, mocks, require.New(t)
}
