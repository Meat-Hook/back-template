package rpc

import (
	"context"
	"errors"

	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/rpc/pb"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/app"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// User handler for getting user info to another microservice.
func (a *api) User(ctx context.Context, req *pb.RequestUser) (*pb.UserInfo, error) {
	info, err := a.app.UserByID(ctx, app.Session{}, int(req.Id))
	if err != nil {
		return nil, apiError(err)
	}

	return apiUser(info), nil
}

func apiUser(user *app.User) *pb.UserInfo {
	return &pb.UserInfo{
		Id:    int64(user.ID),
		Name:  user.Name,
		Email: user.Email,
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
