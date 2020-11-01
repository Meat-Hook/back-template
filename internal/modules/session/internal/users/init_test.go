package users_test

import (
	"testing"

	"github.com/Meat-Hook/back-template/internal/modules/session/internal/users"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func start(t *testing.T) (*users.Client, *MockuserSvc, *require.Assertions) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mock := NewMockuserSvc(ctrl)

	return users.New(mock), mock, require.New(t)
}
