package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"reflect"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/Meat-Hook/back-template/cmd/file/client"
	"github.com/Meat-Hook/back-template/cmd/file/internal/api/rpc"
	"github.com/Meat-Hook/back-template/cmd/file/internal/app"
	"github.com/Meat-Hook/back-template/libs/log"
	librpc "github.com/Meat-Hook/back-template/libs/rpc"
	pb "github.com/Meat-Hook/back-template/proto/gen/go/file/v1"
)

var (
	reqID = xid.New()
	ctx   = log.ReqIDWithCtx(context.Background(), reqID.String())

	errAny = errors.New("any err")
)

func start(t *testing.T, fileID uuid.UUID, fileMD json.RawMessage, file []byte) (*client.Client, *serverMock, *require.Assertions) {
	t.Helper()
	r := require.New(t)

	mock := &serverMock{
		assert:   r,
		fileID:   fileID,
		metadata: fileMD,
		file:     file,
	}

	srv := grpc.NewServer()
	pb.RegisterFileServiceServer(srv, mock)
	ln, err := net.Listen("tcp", "")
	r.Nil(err)
	go func() { r.Nil(srv.Serve(ln)) }()
	t.Cleanup(srv.Stop)

	conn, err := librpc.Client(ctx, ln.Addr().String())
	r.Nil(err)

	svc := client.New(conn)

	return svc, mock, r
}

var _ pb.FileServiceServer = &serverMock{}

type serverMock struct {
	assert   *require.Assertions
	fileID   uuid.UUID
	metadata json.RawMessage
	file     []byte
}

func (s serverMock) Upload(stream pb.FileService_UploadServer) error {
	res, err := ioutil.ReadAll(rpc.NewReader(stream))
	s.assert.Nil(err)

	s.assert.Equal(s.file, res)

	return stream.SendAndClose(&pb.UploadResponse{
		FileId: &pb.UUID{
			Value: s.fileID.String(),
		},
	})
}

func (s serverMock) SetMetadata(_ context.Context, request *pb.SetMetadataRequest) (*pb.SetMetadataResponse, error) {
	fileID, err := uuid.FromString(request.FileId.Value)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, app.ErrNotValidID.Error())
	}

	if s.fileID != fileID {
		return nil, status.Error(codes.NotFound, app.ErrNotFound.Error())
	}

	fileMD, err := json.Marshal(request.Metadata.Details)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	fmt.Println(string(fileMD))

	if !reflect.DeepEqual(s.metadata, json.RawMessage(fileMD)) {
		return nil, status.Error(codes.Internal, "not expected metadata")
	}

	return &pb.SetMetadataResponse{Empty: &emptypb.Empty{}}, nil
}

func (s serverMock) Delete(_ context.Context, request *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	fileID, err := uuid.FromString(request.FileId.Value)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, app.ErrNotValidID.Error())
	}

	if s.fileID != fileID {
		return nil, status.Error(codes.NotFound, app.ErrNotFound.Error())
	}

	return &pb.DeleteResponse{Empty: &emptypb.Empty{}}, nil
}
