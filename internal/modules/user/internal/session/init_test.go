package session_test

import (
	"testing"

	"github.com/Meat-Hook/back-template/internal/modules/user/internal/session"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func start(t *testing.T) (*session.Client, *MocksessionSvc, *require.Assertions) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mock := NewMocksessionSvc(ctrl)

	return session.New(mock), mock, require.New(t)
}
