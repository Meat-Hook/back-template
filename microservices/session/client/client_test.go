package client_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Meat-Hook/back-template/libs/log"
	"github.com/Meat-Hook/back-template/microservices/session/client"
	pb "github.com/Meat-Hook/back-template/proto/go/session/v1"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var (
	_ gomock.Matcher = &protoMatcher{}
	_ gomock.Matcher = &reqIDMatcher{}
)

type protoMatcher struct {
	value proto.Message
}

// Matches for implements gomock.Matcher.
func (p protoMatcher) Matches(x interface{}) bool {
	return proto.Equal(p.value, x.(proto.Message))
}

// String for implements gomock.Matcher.
func (p protoMatcher) String() string {
	return fmt.Sprintf("%v", p.value)
}

type reqIDMatcher struct {
	expect string
}

// Matches for implements gomock.Matcher.
func (r reqIDMatcher) Matches(x interface{}) bool {
	ctx, ok := x.(context.Context)
	if !ok {
		return false
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}

	reqID := strings.Join(md.Get(log.ReqID), "")

	return r.expect == reqID
}

// String for implements gomock.Matcher.
func (r reqIDMatcher) String() string {
	return r.expect
}

func TestClient_Access(t *testing.T) {
	t.Parallel()

	conn, mock, assert := start(t)

	var (
		session = &client.Session{
			ID:     "sessionID",
			UserID: uuid.Must(uuid.NewV4()),
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
	mock.EXPECT().Session(reqIDMatcher{expect: reqID.String()}, protoMatcher{value: &pb.SessionRequest{
		Token: token,
	}}).Return(&pb.SessionResponse{
		Id:     session.ID,
		UserId: session.UserID.String(),
	}, nil)

	// err any
	mock.EXPECT().Session(reqIDMatcher{expect: reqID.String()}, protoMatcher{value: &pb.SessionRequest{
		Token: notValidToken,
	}}).Return(nil, errAny)

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			res, err := conn.Session(ctx, tc.token)
			assert.Equal(tc.want, res)
			assert.ErrorIs(err, tc.wantErr)
		})
	}
}
