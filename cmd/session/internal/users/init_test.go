package users_test

import (
	"testing"

	users2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/users"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func start(t *testing.T) (*users2.Client, *MockuserSvc, *require.Assertions) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mock := NewMockuserSvc(ctrl)

	return users2.New(mock), mock, require.New(t)
}
