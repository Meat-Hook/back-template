package rpc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/rpc/pb"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/app"
	"github.com/golang/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errAny = errors.New("any err")
	user   = app.User{
		ID:    1,
		Email: "username",
		Name:  "email@email.com",
	}

	rpcUser = pb.UserInfo{
		Id:    int64(user.ID),
		Name:  user.Name,
		Email: user.Email,
	}
)

func TestService_GetUserByAuthToken(t *testing.T) {
	t.Parallel()

	c, mockApp, assert := start(t)

	errNotFound := status.Error(codes.NotFound, app.ErrNotFound.Error())
	errDeadline := status.Error(codes.DeadlineExceeded, context.DeadlineExceeded.Error())
	errCanceled := status.Error(codes.Canceled, context.Canceled.Error())
	errInternal := status.Error(codes.Internal, errAny.Error())

	const (
		email = `email@mail.com`
		pass  = `pass`
	)

	testCases := map[string]struct {
		user    *app.User
		want    *pb.UserInfo
		appErr  error
		wantErr error
	}{
		"success":   {&user, &rpcUser, nil, nil},
		"not_found": {nil, nil, app.ErrNotFound, errNotFound},
		"deadline":  {nil, nil, context.DeadlineExceeded, errDeadline},
		"canceled":  {nil, nil, context.Canceled, errCanceled},
		"internal":  {nil, nil, errAny, errInternal},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
	defer cancel()

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			mockApp.EXPECT().Access(gomock.Any(), email, pass).Return(tc.user, tc.appErr)

			res, err := c.Access(ctx, &pb.RequestAccess{
				Email:    email,
				Password: pass,
			})
			if err != nil {
				assert.Equal(tc.wantErr.Error(), err.Error())
			} else {
				assert.Equal(tc.want.Id, res.Id)
				assert.Equal(tc.want.Email, res.Email)
				assert.Equal(tc.want.Name, res.Name)
			}
		})
	}
}
