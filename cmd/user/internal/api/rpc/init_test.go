package rpc_test

import (
	"context"
	"net"
	"os"
	"testing"

	rpc3 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/api/rpc"
	metrics2 "github.com/Meat-Hook/back-template/internal/libs/metrics"
	rpc2 "github.com/Meat-Hook/back-template/internal/libs/rpc"
	pb "github.com/Meat-Hook/back-template/proto/gen/go/user/v1"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestMain(m *testing.M) {
	metrics2.InitMetrics()

	os.Exit(m.Run())
}

func start(t *testing.T) (pb.UserServiceClient, *Mockusers, *require.Assertions) {
	t.Helper()
	r := require.New(t)

	ctrl := gomock.NewController(t)
	mockApp := NewMockusers(ctrl)
	t.Cleanup(ctrl.Finish)

	server := rpc3.New(mockApp, rpc2.Server(zerolog.New(os.Stdout)))

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

	return pb.NewUserServiceClient(conn), mockApp, r
}
