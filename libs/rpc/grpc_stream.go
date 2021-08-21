package rpc

import (
	"context"
	"errors"
	"strings"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"github.com/Meat-Hook/back-template/libs/log"
	"github.com/Meat-Hook/back-template/libs/metrics"
)

// Errors.
var (
	ErrWithoutMD = errors.New("caller without metadata")
)

// MakeStreamServerLogger returns a new stream server interceptor that contains request logger.
func MakeStreamServerLogger(logger zerolog.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return ErrWithoutMD
		}

		logger = newRPCLogger(ctx, logger, info.FullMethod)
		reqID := xid.New().String()
		if res := md.Get(log.ReqID); res != nil {
			reqID = strings.Join(res, "")
		}
		logger = logger.With().Str(log.ReqID, reqID).Logger()

		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = logger.WithContext(log.ReqIDWithCtx(ctx, reqID))

		return handler(srv, wrapped)
	}
}

// MakeStreamServerRecover returns a new stream server interceptor that recover and logs panic.
func MakeStreamServerRecover() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if p := recover(); p != nil {
				metrics.PanicsTotal.Inc()
				l := zerolog.Ctx(stream.Context())
				l.Error().Stack().
					Uint32(log.Code, uint32(codes.Internal)).
					Interface(log.PanicReason, p).Stack().Msg("panic")
				err = errInternal
			}
		}()

		return handler(srv, stream)
	}
}

// StreamServerAccessLog returns a new stream server interceptor that logs request status.
func StreamServerAccessLog(srv interface{}, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	logger := zerolog.Ctx(stream.Context())
	logger.Info().Msg("started")
	defer logger.Info().Msg("finished")

	err = handler(srv, stream)
	err = rpcLogHandler(logger, err)

	return err
}

// MakeStreamClientLogger returns a new stream client interceptor that contains request logger.
func MakeStreamClientLogger(ctx context.Context, desc *grpc.StreamDesc,
	cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	logger := *zerolog.Ctx(ctx)
	logger = newRPCLogger(ctx, logger, method)
	ctx = logger.WithContext(ctx)

	return streamer(ctx, desc, cc, method, opts...)
}

// StreamClientAccessLog returns a new stream client interceptor that logs request status.
func StreamClientAccessLog(ctx context.Context, desc *grpc.StreamDesc,
	cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	logger := zerolog.Ctx(ctx)
	clientStream, err := streamer(ctx, desc, cc, method, opts...)
	err = rpcLogHandler(logger, err)

	return clientStream, err
}
