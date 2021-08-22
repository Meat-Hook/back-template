package rpc

import (
	"context"
	"errors"
	"net"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/Meat-Hook/back-template/cmd/session/internal/app"
	pb "github.com/Meat-Hook/back-template/proto/gen/go/session/v1"
)

// Session implements pb.ServiceServer.
func (a *api) Session(ctx context.Context, request *pb.SessionRequest) (*pb.SessionResponse, error) {
	session, err := a.app.Session(ctx, request.Token)
	if err != nil {
		return nil, apiError(err)
	}

	return &pb.SessionResponse{
		SessionId: &pb.UUID{Value: session.ID.String()},
		UserId:    &pb.UUID{Value: session.UserID.String()},
	}, nil
}

// RemoveSession implements pb.ServiceServer.
func (a *api) RemoveSession(ctx context.Context, request *pb.RemoveSessionRequest) (*pb.RemoveSessionResponse, error) {
	uid, err := uuid.FromString(request.SessionId.Value)
	if err != nil {
		return nil, apiError(err)
	}

	err = a.app.RemoveSession(ctx, uid)
	if err != nil {
		return nil, apiError(err)
	}

	return &pb.RemoveSessionResponse{Empty: &emptypb.Empty{}}, nil
}

// NewSession implements pb.ServiceServer.
func (a *api) NewSession(ctx context.Context, request *pb.NewSessionRequest) (*pb.NewSessionResponse, error) {
	userID, err := uuid.FromString(request.UserId.Value)
	if err != nil {
		return nil, apiError(err)
	}

	token, err := a.app.NewSession(ctx, userID, app.Origin{
		IP:        net.ParseIP(request.Ip),
		UserAgent: request.UserAgent,
	})
	if err != nil {
		return nil, apiError(err)
	}

	return &pb.NewSessionResponse{Token: token.Value}, nil
}

func apiError(err error) error {
	if err == nil {
		return nil
	}

	code := codes.Internal
	switch {
	case errors.Is(err, app.ErrNotFound):
		code = codes.NotFound
	case errors.Is(err, app.ErrInvalidToken):
		code = codes.InvalidArgument
	case errors.Is(err, context.DeadlineExceeded):
		code = codes.DeadlineExceeded
	case errors.Is(err, context.Canceled):
		code = codes.Canceled
	}

	return status.Error(code, err.Error())
}
