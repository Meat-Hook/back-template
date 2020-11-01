package client_test

import (
	"errors"
	"testing"

	"github.com/Meat-Hook/back-template/internal/modules/user/client"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/rpc/pb"
	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ gomock.Matcher = &protoMatcher{}

type protoMatcher struct {
	value proto.Message
}

func (p protoMatcher) Matches(x interface{}) bool {
	return proto.Equal(p.value, x.(proto.Message))
}

func (p protoMatcher) String() string {
	return p.value.String()
}

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
	mock.EXPECT().Access(gomock.Any(), protoMatcher{value: &pb.RequestAccess{
		Email:    user.Email,
		Password: pass,
	}}).Return(&pb.UserInfo{
		Id:    int64(user.ID),
		Name:  user.Name,
		Email: user.Email,
	}, nil)

	// err_any
	mock.EXPECT().
		Access(gomock.Any(), protoMatcher{value: &pb.RequestAccess{}}).
		Return(nil, errAny)

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
