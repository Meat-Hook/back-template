package client_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	client2 "github.com/Meat-Hook/back-template/internal/cmd/user/client"
	log2 "github.com/Meat-Hook/back-template/internal/libs/log"
	pb "github.com/Meat-Hook/back-template/proto/gen/go/user/v1"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var _ gomock.Matcher = &protoMatcher{}

type protoMatcher struct {
	value proto.Message
}

func (p protoMatcher) Matches(x interface{}) bool {
	return proto.Equal(p.value, x.(proto.Message))
}

func (p protoMatcher) String() string {
	return fmt.Sprintf("%v", p.value.ProtoReflect())
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

	reqID := strings.Join(md.Get(log2.ReqID), "")

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
		user = &client2.User{
			ID:    uuid.Must(uuid.NewV4()),
			Email: "email@mail.com",
			Name:  "username",
		}
		pass = `pass`
	)

	testCases := map[string]struct {
		email, pass string
		want        *client2.User
		wantErr     error
	}{
		"success": {user.Email, pass, user, nil},
		"err_any": {"", "", nil, status.Error(codes.Unknown, errAny.Error())},
	}

	// success
	mock.EXPECT().Access(reqIDMatcher{expect: reqID.String()}, protoMatcher{value: &pb.AccessRequest{
		Email:    user.Email,
		Password: pass,
	}}).Return(&pb.AccessResponse{
		Id:    user.ID.String(),
		Name:  user.Name,
		Email: user.Email,
	}, nil)

	// err_any
	mock.EXPECT().
		Access(reqIDMatcher{expect: reqID.String()}, protoMatcher{value: &pb.AccessRequest{}}).
		Return(nil, errAny)

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			res, err := conn.Access(ctx, tc.email, tc.pass)

			assert.ErrorIs(err, tc.wantErr)
			assert.Equal(tc.want, res)
		})
	}
}
