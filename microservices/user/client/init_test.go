package client_test

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/Meat-Hook/back-template/libs/log"
	librpc "github.com/Meat-Hook/back-template/libs/rpc"
	"github.com/Meat-Hook/back-template/microservices/user/client"
	pb "github.com/Meat-Hook/back-template/proto/go/user/v1"
	"github.com/golang/mock/gomock"
	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

//go:generate mockgen -source=../../../proto/go/user/v1/user_grpc.pb.go -destination mock.app.contracts_test.go -package client_test

var (
	reqID = xid.New()
	ctx   = log.ReqIDWithCtx(context.Background(), reqID.String())

	errAny = errors.New("any err")
)

func start(t *testing.T) (*client.Client, *MockUserServiceServer, *require.Assertions) {
	t.Helper()
	r := require.New(t)

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	mock := NewMockUserServiceServer(ctrl)

	srv := grpc.NewServer()
	pb.RegisterUserServiceServer(srv, mock)
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
