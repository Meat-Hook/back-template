package session_test

import (
	"testing"

	"github.com/Meat-Hook/back-template/cmd/user/internal/session"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func start(t *testing.T) (*session.Client, *MocksessionSvc, *require.Assertions) {
	t.Helper()

	ctrl := gomock.NewController(t)

	mock := NewMocksessionSvc(ctrl)

	return session.New(mock), mock, require.New(t)
}
