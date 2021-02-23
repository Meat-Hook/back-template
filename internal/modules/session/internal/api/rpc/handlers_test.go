package rpc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Meat-Hook/back-template/internal/modules/session/internal/api/rpc/pb"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/app"
	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errAny      = errors.New("any err")
	sessionInfo = app.Session{
		ID:     "id",
		UserID: 1,
	}

	rpcUser = pb.SessionInfo{
		ID:     "id",
		UserID: 1,
	}
)

func TestService_GetUserByAuthToken(t *testing.T) {
	t.Parallel()

	errNotFound := status.Error(codes.NotFound, app.ErrNotFound.Error())
	errDeadline := status.Error(codes.DeadlineExceeded, context.DeadlineExceeded.Error())
	errCanceled := status.Error(codes.Canceled, context.Canceled.Error())
	errInternal := status.Error(codes.Internal, errAny.Error())

	const token = `accessToken`

	testCases := map[string]struct {
		session *app.Session
		want    *pb.SessionInfo
		appErr  error
		wantErr error
	}{
		"success":   {&sessionInfo, &rpcUser, nil, nil},
		"not_found": {nil, nil, app.ErrNotFound, errNotFound},
		"deadline":  {nil, nil, context.DeadlineExceeded, errDeadline},
		"canceled":  {nil, nil, context.Canceled, errCanceled},
		"internal":  {nil, nil, errAny, errInternal},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			c, mockApp, assert := start(t)

			mockApp.EXPECT().Session(gomock.Any(), token).Return(tc.session, tc.appErr)

			res, err := c.Session(ctx, &pb.RequestSession{Token: token})
			assert.ErrorIs(err, tc.wantErr)
			assert.True(proto.Equal(tc.want, res))
		})
	}
}
