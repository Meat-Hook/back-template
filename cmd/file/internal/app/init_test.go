package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Meat-Hook/back-template/cmd/file/internal/app"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

var (
	ctx    = context.Background()
	errAny = errors.New("any error")
)

type mocks struct {
	repo *MockRepo
}

func start(t *testing.T) (*app.Module, *mocks, *require.Assertions) {
	t.Helper()
	ctrl := gomock.NewController(t)

	mockRepo := NewMockRepo(ctrl)

	module := app.New(mockRepo)

	mocks := &mocks{
		repo: mockRepo,
	}

	return module, mocks, require.New(t)
}
