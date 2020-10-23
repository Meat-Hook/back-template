package middleware

import (
	"github.com/Meat-Hook/back-template/internal/libs/log"
	"github.com/Meat-Hook/back-template/internal/libs/metrics"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MakeStreamServerLogger returns a new stream server interceptor that contains request logger.
func MakeStreamServerLogger(logBuilder zerolog.Context) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		l := newRPCLogger(ctx, logBuilder, info.FullMethod)
		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = l.WithContext(ctx)

		return handler(srv, wrapped)
	}
}

// StreamServerRecover returns a new stream server interceptor that recover and logs panic.
func StreamServerRecover(srv interface{}, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	defer func() {
		if p := recover(); p != nil {
			metrics.PanicsTotal.Inc()
			l := zerolog.Ctx(stream.Context())
			l.Error().
				Uint32(log.Code, uint32(codes.Internal)).
				Interface(log.Err, p).Stack().Msg("panic")

			err = status.Errorf(codes.Internal, "%v", p)
		}
	}()

	return handler(srv, stream)
}

// StreamServerAccessLog returns a new stream server interceptor that logs request status.
func StreamServerAccessLog(srv interface{}, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	err = handler(srv, stream)
	l := zerolog.Ctx(stream.Context())
	RPCLogHandler(l, err)

	return err
}
