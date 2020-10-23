package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Meat-Hook/back-template/internal/modules/user/internal/app"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

var (
	ctx    = context.Background()
	errAny = errors.New("any error")
)

type mocks struct {
	hasher       *mock.MockHasher
	repo         *mock.MockRepo
	code         *mock.MockCode
	notification *mock.MockNotification
	auth         *mock.MockAuth
}

func start(t *testing.T) (*app.Module, *mocks, *require.Assertions) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockRepo := mock.NewMockRepo(ctrl)
	mockHasher := mock.NewMockHasher(ctrl)
	mockCode := mock.NewMockCode(ctrl)
	mockNotification := mock.NewMockNotification(ctrl)
	mockAuth := mock.NewMockAuth(ctrl)

	module := app.New(mockRepo, mockHasher, mockNotification, mockCode, mockAuth)

	mocks := &mocks{
		hasher:       mockHasher,
		repo:         mockRepo,
		code:         mockCode,
		notification: mockNotification,
		auth:         mockAuth,
	}

	return module, mocks, require.New(t)
}
