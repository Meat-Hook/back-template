package rpc

import (
	"context"
	"errors"

	"github.com/Meat-Hook/back-template/cmd/session/internal/app"
	pb "github.com/Meat-Hook/back-template/proto/gen/go/session/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Session get user session by raw token.
func (a *api) Session(ctx context.Context, req *pb.SessionRequest) (*pb.SessionResponse, error) {
	info, err := a.app.Session(ctx, req.Token)
	if err != nil {
		return nil, apiError(err)
	}

	return apiSession(info), nil
}

func apiSession(session *app.Session) *pb.SessionResponse {
	return &pb.SessionResponse{
		Id:     session.ID,
		UserId: session.UserID.String(),
	}
}

func apiError(err error) error {
	if err == nil {
		return nil
	}

	code := codes.Internal
	switch {
	case errors.Is(err, app.ErrNotFound):
		code = codes.NotFound
	case errors.Is(err, context.DeadlineExceeded):
		code = codes.DeadlineExceeded
	case errors.Is(err, context.Canceled):
		code = codes.Canceled
	}

	return status.Error(code, err.Error())
}
