package rpc_test

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/Meat-Hook/back-template/cmd/file/internal/api/rpc"
	"github.com/Meat-Hook/back-template/libs/metrics"
	librpc "github.com/Meat-Hook/back-template/libs/rpc"
	pb "github.com/Meat-Hook/back-template/proto/gen/go/file/v1"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

const testFile = `test.jpg`

var (
	fileID = uuid.Must(uuid.NewV4())
)

func TestMain(m *testing.M) {
	metrics.InitMetrics()

	os.Exit(m.Run())
}

func start(t *testing.T) (pb.FileServiceClient, *Mockfiles, *require.Assertions) {
	t.Helper()
	r := require.New(t)

	ctrl := gomock.NewController(t)
	mockApp := NewMockfiles(ctrl)

	server := rpc.New(mockApp, librpc.Server(zerolog.New(os.Stdout)))

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	r.Nil(err)

	go func() {
		err := server.Serve(ln)
		r.Nil(err)
	}()

	ctx, cancel := context.WithCancel(context.Background())
	conn, err := grpc.DialContext(ctx, ln.Addr().String(),
		grpc.WithInsecure(), // TODO Add TLS and remove this.
		grpc.WithBlock(),
	)
	r.Nil(err)

	t.Cleanup(func() {
		err := conn.Close()
		r.Nil(err)
		server.GracefulStop()
		cancel()
	})

	return pb.NewFileServiceClient(conn), mockApp, r
}
