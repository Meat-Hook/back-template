package client_test

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/Meat-Hook/back-template/cmd/session/client"
	"github.com/Meat-Hook/back-template/libs/log"
	pb "github.com/Meat-Hook/back-template/proto/gen/go/session/v1"
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

func TestClient_Session(t *testing.T) {
	t.Parallel()

	var (
		session = &client.Session{
			ID:     uuid.Must(uuid.NewV4()),
			UserID: uuid.Must(uuid.NewV4()),
		}
		token             = `token`
		notValidToken     = `notValidToken`
		internalStatusErr = status.Error(codes.Internal, errAny.Error())
	)

	testCases := map[string]struct {
		token       string
		appResponse *pb.SessionResponse
		appError    error
		want        *client.Session
		wantErr     error
	}{
		"success":   {token, &pb.SessionResponse{SessionId: &pb.UUID{Value: session.ID.String()}, UserId: &pb.UUID{Value: session.UserID.String()}}, nil, session, nil},
		"not_found": {notValidToken, nil, status.Error(codes.NotFound, "not found"), nil, client.ErrNotFound},
		"err_any":   {notValidToken, nil, internalStatusErr, nil, internalStatusErr},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			conn, mock, assert := start(t)

			mock.EXPECT().Session(reqIDMatcher{expect: reqID.String()}, protoMatcher{value: &pb.SessionRequest{Token: tc.token}}).
				Return(tc.appResponse, tc.appError)

			res, err := conn.Session(ctx, tc.token)
			assert.Equal(tc.want, res)
			assert.ErrorIs(err, tc.wantErr)
		})
	}
}

func TestClient_RemoveSession(t *testing.T) {
	t.Parallel()

	var (
		internalStatusErr = status.Error(codes.Internal, errAny.Error())
		sessionID         = uuid.Must(uuid.NewV4())
	)

	testCases := map[string]struct {
		appResponse *pb.RemoveSessionResponse
		appError    error
		wantErr     error
	}{
		"success": {&pb.RemoveSessionResponse{Empty: &emptypb.Empty{}}, nil, nil},
		"err_any": {nil, internalStatusErr, status.Error(codes.Internal, errAny.Error())},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			conn, mock, assert := start(t)

			mock.EXPECT().RemoveSession(reqIDMatcher{expect: reqID.String()}, protoMatcher{value: &pb.RemoveSessionRequest{SessionId: &pb.UUID{Value: sessionID.String()}}}).
				Return(tc.appResponse, tc.appError)

			err := conn.RemoveSession(ctx, sessionID)
			assert.ErrorIs(err, tc.wantErr)
		})
	}
}

func TestClient_NewSession(t *testing.T) {
	t.Parallel()

	var (
		internalStatusErr = status.Error(codes.Internal, errAny.Error())
		userID            = uuid.Must(uuid.NewV4())
		ip                = net.ParseIP("192.100.10.4")
		userAgent         = "userAgent"
		token             = "token"
	)

	testCases := map[string]struct {
		appResponse *pb.NewSessionResponse
		appError    error
		want        *client.Token
		wantErr     error
	}{
		"success": {&pb.NewSessionResponse{Token: token}, nil, &client.Token{Value: token}, nil},
		"err_any": {nil, internalStatusErr, nil, status.Error(codes.Internal, errAny.Error())},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			conn, mock, assert := start(t)

			mock.EXPECT().NewSession(reqIDMatcher{expect: reqID.String()}, protoMatcher{value: &pb.NewSessionRequest{
				UserId:    &pb.UUID{Value: userID.String()},
				Ip:        ip.String(),
				UserAgent: userAgent,
			}}).Return(tc.appResponse, tc.appError)

			token, err := conn.NewSession(ctx, userID, ip, userAgent)
			assert.ErrorIs(err, tc.wantErr)
			assert.Equal(tc.want, token)
		})
	}
}
