package rpc_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/Meat-Hook/back-template/cmd/session/internal/app"
	pb "github.com/Meat-Hook/back-template/proto/gen/go/session/v1"
)

var (
	errAny = errors.New("any err")
	origin = app.Origin{
		IP:        net.ParseIP("192.100.10.4"),
		UserAgent: "UserAgent",
	}
)

func TestApi_Session(t *testing.T) {
	t.Parallel()

	sessionInfo := app.Session{
		ID:     uuid.Must(uuid.NewV4()),
		Origin: origin,
		Token: app.Token{
			Value: "token",
		},
		UserID:    uuid.Must(uuid.NewV4()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	sessionResponse := pb.SessionResponse{
		SessionId: &pb.UUID{Value: sessionInfo.ID.String()},
		UserId:    &pb.UUID{Value: sessionInfo.UserID.String()},
	}

	errNotFound := status.Error(codes.NotFound, app.ErrNotFound.Error())
	errDeadline := status.Error(codes.DeadlineExceeded, context.DeadlineExceeded.Error())
	errCanceled := status.Error(codes.Canceled, context.Canceled.Error())
	errInternal := status.Error(codes.Internal, errAny.Error())

	const token = `accessToken`

	testCases := []struct {
		name    string
		session *app.Session
		want    *pb.SessionResponse
		appErr  error
		wantErr error
	}{
		{"success", &sessionInfo, &sessionResponse, nil, nil},
		{"err_not_found", nil, nil, app.ErrNotFound, errNotFound},
		{"err_deadline", nil, nil, context.DeadlineExceeded, errDeadline},
		{"err_canceled", nil, nil, context.Canceled, errCanceled},
		{"err_any", nil, nil, errAny, errInternal},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			c, mockApp, assert := start(t, prometheus.NewPedanticRegistry())

			mockApp.EXPECT().Session(gomock.Any(), token).Return(tc.session, tc.appErr)

			res, err := c.Session(ctx, &pb.SessionRequest{Token: token})
			assert.ErrorIs(err, tc.wantErr)
			assert.True(proto.Equal(tc.want, res))
		})
	}
}

func TestApi_RemoveSession(t *testing.T) {
	t.Parallel()

	errNotFound := status.Error(codes.NotFound, app.ErrNotFound.Error())
	errDeadline := status.Error(codes.DeadlineExceeded, context.DeadlineExceeded.Error())
	errCanceled := status.Error(codes.Canceled, context.Canceled.Error())
	errInternal := status.Error(codes.Internal, errAny.Error())

	testCases := []struct {
		name   string
		appErr error
		want   error
	}{
		{"success", nil, nil},
		{"err_not_found", app.ErrNotFound, errNotFound},
		{"err_deadline", context.DeadlineExceeded, errDeadline},
		{"err_canceled", context.Canceled, errCanceled},
		{"err_any", errAny, errInternal},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			sessionID := uuid.Must(uuid.NewV4())

			c, mockApp, assert := start(t, prometheus.NewPedanticRegistry())

			mockApp.EXPECT().RemoveSession(gomock.Any(), sessionID).Return(tc.appErr)

			_, err := c.RemoveSession(ctx, &pb.RemoveSessionRequest{SessionId: &pb.UUID{Value: sessionID.String()}})
			assert.ErrorIs(err, tc.want)
		})
	}
}

func TestApi_NewSession(t *testing.T) {
	t.Parallel()

	errNotFound := status.Error(codes.NotFound, app.ErrNotFound.Error())
	errDeadline := status.Error(codes.DeadlineExceeded, context.DeadlineExceeded.Error())
	errCanceled := status.Error(codes.Canceled, context.Canceled.Error())
	errInternal := status.Error(codes.Internal, errAny.Error())

	const token = `token`

	testCases := []struct {
		name     string
		appToken *app.Token
		want     *pb.NewSessionResponse
		appErr   error
		wantErr  error
	}{
		{"success", &app.Token{Value: token}, &pb.NewSessionResponse{Token: token}, nil, nil},
		{"err_not_found", nil, nil, app.ErrNotFound, errNotFound},
		{"err_deadline", nil, nil, context.DeadlineExceeded, errDeadline},
		{"err_canceled", nil, nil, context.Canceled, errCanceled},
		{"err_any", nil, nil, errAny, errInternal},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			userID := uuid.Must(uuid.NewV4())

			c, mockApp, assert := start(t, prometheus.NewPedanticRegistry())

			mockApp.EXPECT().NewSession(gomock.Any(), userID, origin).Return(tc.appToken, tc.appErr)

			res, err := c.NewSession(ctx, &pb.NewSessionRequest{
				UserId:    &pb.UUID{Value: userID.String()},
				Ip:        origin.IP.String(),
				UserAgent: origin.UserAgent,
			})
			assert.ErrorIs(err, tc.wantErr)
			assert.True(proto.Equal(tc.want, res))
		})
	}
}
