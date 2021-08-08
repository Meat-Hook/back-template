package session_test

import (
	"context"
	"errors"
	"net"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/Meat-Hook/back-template/cmd/user/internal/app"
	"github.com/Meat-Hook/back-template/cmd/user/internal/services/session"
	"github.com/Meat-Hook/back-template/libs/metrics"
)

var (
	reg = prometheus.NewPedanticRegistry()

	ctx    = context.Background()
	errAny = errors.New("any err")
	origin = app.Origin{
		IP:        net.ParseIP("192.100.10.4"),
		UserAgent: "UserAgent",
	}
)

func TestMain(m *testing.M) {
	metrics.InitMetrics(reg)

	os.Exit(m.Run())
}

func start(t *testing.T) (*session.Client, *MocksessionSvc, *require.Assertions) {
	t.Helper()

	ctrl := gomock.NewController(t)

	mock := NewMocksessionSvc(ctrl)

	return session.New(mock), mock, require.New(t)
}
