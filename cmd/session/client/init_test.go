package client_test

import (
	"context"
	"errors"
	"net"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/Meat-Hook/back-template/cmd/session/client"
	"github.com/Meat-Hook/back-template/libs/log"
	"github.com/Meat-Hook/back-template/libs/metrics"
	"github.com/Meat-Hook/back-template/libs/rpc"
	pb "github.com/Meat-Hook/back-template/proto/gen/go/session/v1"
)

var (
	logger = zerolog.New(os.Stdout)
	reqID  = xid.New()
	ctx    = log.ReqIDWithCtx(context.Background(), reqID.String())

	errAny       = errors.New("any err")
	reg          = prometheus.NewPedanticRegistry()
	clientMetric = rpc.NewClientMetrics(reg, "test")
)

func TestMain(m *testing.M) {
	metrics.InitMetrics(reg)

	os.Exit(m.Run())
}

func start(t *testing.T) (*client.Client, *MockServiceServer, *require.Assertions) {
	t.Helper()
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	mock := NewMockServiceServer(ctrl)

	srv := grpc.NewServer()
	pb.RegisterServiceServer(srv, mock)
	ln, err := net.Listen("tcp", "")
	assert.NoError(err)
	go func() { assert.NoError(srv.Serve(ln)) }()

	t.Cleanup(func() {
		srv.Stop()
	})

	conn, err := rpc.Dial(ctx, logger, ln.Addr().String(), clientMetric)
	assert.NoError(err)

	svc := client.New(conn)

	return svc, mock, assert
}
