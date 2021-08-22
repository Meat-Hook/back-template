package app_test

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/Meat-Hook/back-template/cmd/user/internal/app"
)

var (
	ctx    = context.Background()
	errAny = errors.New("any error")
	origin = app.Origin{
		IP:        net.ParseIP("192.100.10.4"),
		UserAgent: "UserAgent",
	}
)

type mocks struct {
	hasher *MockHasher
	repo   *MockRepo
	auth   *MockAuthSvc
	file   *MockFileSvc
}

func start(t *testing.T) (*app.Module, *mocks, *require.Assertions) {
	t.Helper()
	ctrl := gomock.NewController(t)

	mockRepo := NewMockRepo(ctrl)
	mockHasher := NewMockHasher(ctrl)
	mockAuth := NewMockAuthSvc(ctrl)
	mockFile := NewMockFileSvc(ctrl)

	module := app.New(mockRepo, mockHasher, mockAuth, mockFile)

	mocks := &mocks{
		hasher: mockHasher,
		repo:   mockRepo,
		auth:   mockAuth,
		file:   mockFile,
	}

	return module, mocks, require.New(t)
}
