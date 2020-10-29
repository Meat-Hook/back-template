package client_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Meat-Hook/back-template/internal/modules/user/client"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/rpc/pb"
	"github.com/golang/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestClient_Access(t *testing.T) {
	t.Parallel()

	conn, mock, assert := start(t)

	var (
		user = &client.User{
			ID:    1,
			Email: "email@mail.com",
			Name:  "username",
		}
		pass = `pass`
	)

	testCases := map[string]struct {
		email, pass string
		want        *client.User
		wantErr     error
	}{
		"success": {user.Email, pass, user, nil},
		"err_any": {"", "", nil, status.Error(codes.Unknown, errAny.Error())},
	}

	// success
	mock.EXPECT().Access(gomock.Any(), gomock.Any()).
		Do(func(ctx context.Context, arg *pb.RequestAccess) {
			assert.Equal(user.Email, arg.Email)
			assert.Equal(pass, arg.Password)
		}).Return(&pb.UserInfo{
		Id:    int64(user.ID),
		Name:  user.Name,
		Email: user.Email,
	}, nil)

	// err_any
	mock.EXPECT().Access(gomock.Any(), gomock.Any()).
		Do(func(ctx context.Context, arg *pb.RequestAccess) {
			assert.Zero(arg.Email)
			assert.Zero(arg.Password)
		}).Return(nil, errAny)

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {

			res, err := conn.Access(ctx, tc.email, tc.pass)
			if err != nil {
				assert.Equal(tc.wantErr.Error(), errors.Unwrap(err).Error())
			} else {
				assert.Nil(err)
			}
			assert.Equal(tc.want, res)
		})
	}
}
