package users_test

import (
	"testing"

	"github.com/Meat-Hook/back-template/cmd/session/internal/users"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func start(t *testing.T) (*users.Client, *MockuserSvc, *require.Assertions) {
	t.Helper()

	ctrl := gomock.NewController(t)
	mock := NewMockuserSvc(ctrl)

	return users.New(mock), mock, require.New(t)
}
