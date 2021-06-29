package rpc_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/Meat-Hook/back-template/cmd/file/internal/app"
	pb "github.com/Meat-Hook/back-template/proto/gen/go/file/v1"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	errAny = errors.New("any err")
)

var _ gomock.Matcher = &fileMatcher{}

type fileMatcher struct {
	file   []byte
	assert *require.Assertions
}

func (f fileMatcher) Matches(x interface{}) bool {
	argFile, err := io.ReadAll(x.(io.Reader))
	f.assert.Nil(err)

	return bytes.Equal(argFile, f.file)
}

func (f fileMatcher) String() string {
	return "fileMatcher"
}

func TestApi_Upload(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	file, err := os.Open(testFile)
	assert.Nil(err)
	t.Cleanup(func() {
		assert.Nil(file.Close())
	})

	buf, err := io.ReadAll(file)
	assert.Nil(err)

	testCases := map[string]struct {
		file    io.Reader
		want    *pb.UploadResponse
		appRes  uuid.UUID
		appErr  error
		wantErr error
	}{
		"success": {bytes.NewBuffer(buf), &pb.UploadResponse{FileId: &pb.UUID{Value: fileID.String()}}, fileID, nil, nil},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			c, mockApp, assert := start(t)

			mockApp.EXPECT().
				UploadFile(gomock.Any(), fileMatcher{buf, assert}).
				Return(tc.appRes, tc.appErr)

			stream, err := c.Upload(ctx)
			assert.Nil(err)

			buf = make([]byte, app.MaxChunkSize)

			for {
				n, err := tc.file.Read(buf)
				if err != nil && !errors.Is(err, io.EOF) {
					assert.Nil(err)
				}

				if n == 0 {
					break
				}

				in := &pb.UploadRequest{
					Chunk: &pb.Chunk{
						Content: buf[:n],
					},
				}

				err = stream.Send(in)
				assert.Nil(err)
			}

			res, err := stream.CloseAndRecv()
			assert.ErrorIs(err, tc.wantErr)
			assert.True(proto.Equal(tc.want, res))
		})
	}
}

func TestApi_SetMetadata(t *testing.T) {
	t.Parallel()

	errNotFound := status.Error(codes.NotFound, app.ErrNotFound.Error())
	errDeadline := status.Error(codes.DeadlineExceeded, context.DeadlineExceeded.Error())
	errCanceled := status.Error(codes.Canceled, context.Canceled.Error())
	errInternal := status.Error(codes.Internal, errAny.Error())

	md := map[string]interface{}{"key": "value"}
	testCases := map[string]struct {
		fileID uuid.UUID
		appErr error
		want   error
	}{
		"success":       {uuid.Must(uuid.NewV4()), nil, nil},
		"err_not_found": {uuid.Must(uuid.NewV4()), app.ErrNotFound, errNotFound},
		"err_deadline":  {uuid.Must(uuid.NewV4()), context.DeadlineExceeded, errDeadline},
		"err_canceled":  {uuid.Must(uuid.NewV4()), context.Canceled, errCanceled},
		"err_any":       {uuid.Must(uuid.NewV4()), errAny, errInternal},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			c, mockApp, assert := start(t)

			js, err := json.Marshal(md)
			assert.Nil(err)

			mockApp.EXPECT().SetMetadata(gomock.Any(), tc.fileID, js).Return(tc.appErr)

			st, err := structpb.NewStruct(md)
			assert.Nil(err)

			_, err = c.SetMetadata(ctx, &pb.SetMetadataRequest{
				FileId: &pb.UUID{Value: tc.fileID.String()},
				Metadata: &pb.Metadata{
					Details: st,
				},
			})

			assert.ErrorIs(err, tc.want)
		})
	}
}

func TestApi_Delete(t *testing.T) {
	t.Parallel()

	errNotFound := status.Error(codes.NotFound, app.ErrNotFound.Error())
	errDeadline := status.Error(codes.DeadlineExceeded, context.DeadlineExceeded.Error())
	errCanceled := status.Error(codes.Canceled, context.Canceled.Error())
	errInternal := status.Error(codes.Internal, errAny.Error())

	testCases := map[string]struct {
		fileID uuid.UUID
		appErr error
		want   error
	}{
		"success":       {uuid.Must(uuid.NewV4()), nil, nil},
		"err_not_found": {uuid.Must(uuid.NewV4()), app.ErrNotFound, errNotFound},
		"err_deadline":  {uuid.Must(uuid.NewV4()), context.DeadlineExceeded, errDeadline},
		"err_canceled":  {uuid.Must(uuid.NewV4()), context.Canceled, errCanceled},
		"err_any":       {uuid.Must(uuid.NewV4()), errAny, errInternal},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			c, mockApp, assert := start(t)

			mockApp.EXPECT().Delete(gomock.Any(), tc.fileID).Return(tc.appErr)

			_, err := c.Delete(ctx, &pb.DeleteRequest{
				FileId: &pb.UUID{Value: tc.fileID.String()},
			})

			assert.ErrorIs(err, tc.want)
		})
	}
}
