package rpc_test

import (
	"context"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/Meat-Hook/back-template/cmd/file/internal/api/rpc"
	"github.com/Meat-Hook/back-template/libs/metrics"
	pb "github.com/Meat-Hook/back-template/proto/gen/go/file/v1"
)

const testFile = `test.jpg`

var (
	reg    = prometheus.NewPedanticRegistry()
	fileID = uuid.Must(uuid.NewV4())
)

func TestMain(m *testing.M) {
	metrics.InitMetrics(reg)

	os.Exit(m.Run())
}

func start(t *testing.T) (pb.ServiceClient, *Mockfiles, *require.Assertions) {
	t.Helper()
	assert := require.New(t)

	ctrl := gomock.NewController(t)
	mockApp := NewMockfiles(ctrl)
	logger := zerolog.New(os.Stdout)

	server := rpc.New(logger.WithContext(context.Background()), reg, strings.Replace(t.Name(), "/", "_", -1), mockApp)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(err)

	go func() {
		err := server.Serve(ln)
		assert.NoError(err)
	}()

	ctx, cancel := context.WithCancel(context.Background())
	conn, err := grpc.DialContext(ctx, ln.Addr().String(),
		grpc.WithInsecure(), // TODO Add TLS and remove this.
		grpc.WithBlock(),
	)
	assert.NoError(err)

	t.Cleanup(func() {
		err := conn.Close()
		assert.NoError(err)
		server.GracefulStop()
		cancel()
	})

	return pb.NewServiceClient(conn), mockApp, assert
}
