package session_test

import (
	"testing"

	session2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/session"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func start(t *testing.T) (*session2.Client, *MocksessionSvc, *require.Assertions) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mock := NewMocksessionSvc(ctrl)

	return session2.New(mock), mock, require.New(t)
}
