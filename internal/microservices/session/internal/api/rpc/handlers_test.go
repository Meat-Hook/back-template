package rpc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Meat-Hook/back-template/internal/microservices/session/internal/api/rpc/pb"
	"github.com/Meat-Hook/back-template/internal/microservices/session/internal/app"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var (
	errAny      = errors.New("any err")
	sessionInfo = app.Session{
		ID:     "id",
		UserID: uuid.Must(uuid.NewV4()),
	}

	rpcUser = pb.SessionInfo{
		ID:     "id",
		UserID: sessionInfo.UserID.String(),
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
		"success":       {&sessionInfo, &rpcUser, nil, nil},
		"err_not_found": {nil, nil, app.ErrNotFound, errNotFound},
		"err_deadline":  {nil, nil, context.DeadlineExceeded, errDeadline},
		"err_canceled":  {nil, nil, context.Canceled, errCanceled},
		"err_any":       {nil, nil, errAny, errInternal},
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
