package rpc_test

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/Meat-Hook/back-template/internal/libs/metrics"
	librpc "github.com/Meat-Hook/back-template/internal/libs/rpc"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/rpc"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/rpc/pb"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestMain(m *testing.M) {
	metrics.InitMetrics()

	os.Exit(m.Run())
}

func start(t *testing.T) (pb.UserClient, *Mockusers, *require.Assertions) {
	t.Helper()
	r := require.New(t)

	ctrl := gomock.NewController(t)
	mockApp := NewMockusers(ctrl)
	t.Cleanup(ctrl.Finish)

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

	return pb.NewUserClient(conn), mockApp, r
}
