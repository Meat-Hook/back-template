package rpc

import (
	"context"
	"errors"

	"github.com/Meat-Hook/back-template/cmd/file/internal/app"
	pb "github.com/Meat-Hook/back-template/proto/gen/go/file/v1"
	"github.com/gofrs/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Upload file to database.
func (a *api) Upload(stream pb.FileService_UploadServer) error {
	id, err := a.app.UploadFile(stream.Context(), NewReader(stream))
	if err != nil {
		return apiError(err)
	}

	return stream.SendAndClose(&pb.UploadResponse{FileId: &pb.UUID{Value: id.String()}})
}

// SetMetadata for file.
func (a *api) SetMetadata(ctx context.Context, request *pb.SetMetadataRequest) (*pb.SetMetadataResponse, error) {
	id, err := uuid.FromString(request.FileId.Value)
	if err != nil {
		return nil, apiError(app.ErrNotValidID)
	}

	js, err := request.Metadata.Details.MarshalJSON()
	if err != nil {
		return nil, apiError(err)
	}

	err = a.app.SetMetadata(ctx, id, js)
	if err != nil {
		return nil, apiError(err)
	}

	return &pb.SetMetadataResponse{Empty: &emptypb.Empty{}}, nil
}

// Delete file from database.
func (a *api) Delete(ctx context.Context, request *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	id, err := uuid.FromString(request.FileId.Value)
	if err != nil {
		return nil, apiError(app.ErrNotValidID)
	}

	err = a.app.Delete(ctx, id)
	if err != nil {
		return nil, apiError(err)
	}

	return &pb.DeleteResponse{Empty: &emptypb.Empty{}}, nil
}

func apiError(err error) error {
	if err == nil {
		return nil
	}

	code := codes.Internal
	switch {
	case errors.Is(err, app.ErrNotFound):
		code = codes.NotFound
	case errors.Is(err, app.ErrNotValidID):
		code = codes.InvalidArgument
	case errors.Is(err, context.DeadlineExceeded):
		code = codes.DeadlineExceeded
	case errors.Is(err, context.Canceled):
		code = codes.Canceled
	}

	return status.Error(code, err.Error())
}
