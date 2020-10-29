package rpc

import (
	"context"
	"errors"

	"github.com/Meat-Hook/back-template/internal/modules/session/internal/api/rpc/pb"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/app"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Session get user session by raw token.
func (a *api) Session(ctx context.Context, req *pb.RequestSession) (*pb.SessionInfo, error) {
	info, err := a.app.Session(ctx, req.Token)
	if err != nil {
		return nil, apiError(err)
	}

	return apiSession(info), nil
}

func apiSession(session *app.Session) *pb.SessionInfo {
	return &pb.SessionInfo{
		ID:     session.ID,
		UserID: int64(session.UserID),
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
