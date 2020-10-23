package middleware

import (
	"context"

	"github.com/Meat-Hook/back-template/internal/libs/log"
	"github.com/Meat-Hook/back-template/internal/libs/metrics"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MakeUnaryServerLogger returns a new unary server interceptor that contains request logger.
func MakeUnaryServerLogger(logBuilder zerolog.Context) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		l := newRPCLogger(ctx, logBuilder, info.FullMethod)

		return handler(l.WithContext(ctx), req)
	}
}

// UnaryServerRecover returns a new unary server interceptor that recover and logs panic.
func UnaryServerRecover(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
	defer func() {
		if p := recover(); p != nil {
			metrics.PanicsTotal.Inc()

			l := zerolog.Ctx(ctx)
			l.Error().
				Uint32(log.Code, uint32(codes.Internal)).
				Interface(log.Err, p).Stack().Msg("panic")

			err = status.Errorf(codes.Internal, "%v", p)
		}
	}()

	return handler(ctx, req)
}

// UnaryServerAccessLog returns a new unary server interceptor that logs request status.
func UnaryServerAccessLog(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
	resp, err := handler(ctx, req)
	l := zerolog.Ctx(ctx)
	RPCLogHandler(l, err)

	return resp, err
}
