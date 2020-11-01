package client_test

import (
	"context"
	"errors"
	"net"
	"testing"

	librpc "github.com/Meat-Hook/back-template/internal/libs/rpc"
	"github.com/Meat-Hook/back-template/internal/modules/session/client"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/api/rpc/pb"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

//go:generate mockgen -source=../internal/api/rpc/pb/session.pb.go -destination mock.app.contracts_test.go -package client_test

var (
	ctx = context.Background()

	errAny = errors.New("any err")
)

func start(t *testing.T) (*client.Client, *MockSessionServer, *require.Assertions) {
	t.Helper()
	r := require.New(t)

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	mock := NewMockSessionServer(ctrl)

	srv := grpc.NewServer()
	pb.RegisterSessionServer(srv, mock)
	ln, err := net.Listen("tcp", "")
	r.Nil(err)
	go func() { r.Nil(srv.Serve(ln)) }()

	t.Cleanup(func() {
		srv.Stop()
	})

	conn, err := librpc.Client(ctx, ln.Addr().String())
	r.Nil(err)

	svc := client.New(conn)

	return svc, mock, r
}
