package session_test

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	session2 "github.com/Meat-Hook/back-template/cmd/user/internal/services/session"
	"github.com/Meat-Hook/back-template/libs/metrics"
)

var (
	reg = prometheus.NewPedanticRegistry()
)

func TestMain(m *testing.M) {
	metrics.InitMetrics(reg)

	os.Exit(m.Run())
}

func start(t *testing.T) (*session2.Client, *MocksessionSvc, *require.Assertions) {
	t.Helper()

	ctrl := gomock.NewController(t)

	mock := NewMocksessionSvc(ctrl)

	return session2.New(mock), mock, require.New(t)
}
