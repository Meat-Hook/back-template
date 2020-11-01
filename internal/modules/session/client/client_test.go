package client_test

import (
	"errors"
	"testing"

	"github.com/Meat-Hook/back-template/internal/modules/session/client"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/api/rpc/pb"
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
		session = &client.Session{
			ID:     "sessionID",
			UserID: 1,
		}
		token         = `token`
		notValidToken = `notValidToken`
	)

	testCases := map[string]struct {
		token   string
		want    *client.Session
		wantErr error
	}{
		"success": {token, session, nil},
		"err_any": {notValidToken, nil, status.Error(codes.Unknown, errAny.Error())},
	}

	// success
	mock.EXPECT().Session(gomock.Any(), protoMatcher{value: &pb.RequestSession{
		Token: token,
	}}).Return(&pb.SessionInfo{
		ID:     session.ID,
		UserID: int64(session.UserID),
	}, nil)

	// err any
	mock.EXPECT().Session(gomock.Any(), protoMatcher{value: &pb.RequestSession{
		Token: notValidToken,
	}}).Return(nil, errAny)

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {

			res, err := conn.Session(ctx, tc.token)
			if err != nil {
				assert.Equal(tc.wantErr.Error(), errors.Unwrap(err).Error())
			} else {
				assert.Nil(err)
			}
			assert.Equal(tc.want, res)
		})
	}
}
