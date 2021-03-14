package rpc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Meat-Hook/back-template/microservices/user/internal/api/rpc/pb"
	"github.com/Meat-Hook/back-template/microservices/user/internal/app"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var (
	errAny = errors.New("any err")
	user   = app.User{
		ID:    uuid.Must(uuid.NewV4()),
		Email: "username",
		Name:  "email@email.com",
	}

	rpcUser = pb.UserInfo{
		Id:    user.ID.String(),
		Name:  user.Name,
		Email: user.Email,
	}
)

func TestService_GetUserByAuthToken(t *testing.T) {
	t.Parallel()

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
		"success":       {&user, &rpcUser, nil, nil},
		"err_not_found": {nil, nil, app.ErrNotFound, errNotFound},
		"err_deadline":  {nil, nil, context.DeadlineExceeded, errDeadline},
		"err_canceled":  {nil, nil, context.Canceled, errCanceled},
		"err_any":       {nil, nil, errAny, errInternal},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			c, mockApp, assert := start(t)

			mockApp.EXPECT().Access(gomock.Any(), email, pass).Return(tc.user, tc.appErr)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			res, err := c.Access(ctx, &pb.RequestAccess{
				Email:    email,
				Password: pass,
			})
			assert.ErrorIs(err, tc.wantErr)
			assert.True(proto.Equal(tc.want, res))
		})
	}
}
