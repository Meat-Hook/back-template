package rpc

import (
	"context"
	"strings"

	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"github.com/Meat-Hook/back-template/libs/log"
	"github.com/Meat-Hook/back-template/libs/metrics"
)

// MakeUnaryServerLogger returns a new unary server interceptor that contains request logger.
func MakeUnaryServerLogger(logger zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, ErrWithoutMD
		}

		l := newRPCLogger(ctx, logger, info.FullMethod)
		reqID := xid.New().String()
		if res := md.Get(log.ReqID); res != nil {
			reqID = strings.Join(res, "")
		}
		l = l.With().Str(log.ReqID, reqID).Logger()

		return handler(l.WithContext(log.ReqIDWithCtx(ctx, reqID)), req)
	}
}

// MakeUnaryServerRecover returns a new unary server interceptor that recover and logs panic.
func MakeUnaryServerRecover() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		defer func() {
			if p := recover(); p != nil {
				metrics.PanicsTotal.Inc()
				l := zerolog.Ctx(ctx)
				l.Error().Stack().
					Uint32(log.Code, uint32(codes.Internal)).
					Interface(log.PanicReason, p).Stack().Msg("panic")
				err = errInternal
			}
		}()
		res, err := handler(ctx, req)

		return res, err
	}
}

// UnaryServerAccessLog returns a new unary server interceptor that logs request status.
func UnaryServerAccessLog(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
	resp, err := handler(ctx, req)
	l := zerolog.Ctx(ctx)
	err = rpcLogHandler(l, err)

	return resp, err
}

// MakeUnaryClientLogger returns a new unary client interceptor that contains request logger.
func MakeUnaryClientLogger(logger zerolog.Logger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		l := newRPCLogger(ctx, logger, method)
		ctx = l.WithContext(ctx)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// UnaryClientAccessLog returns a new unary client interceptor that logs request status.
func UnaryClientAccessLog(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	err := invoker(ctx, method, req, reply, cc, opts...)
	l := zerolog.Ctx(ctx)
	err = rpcLogHandler(l, err)
	return err
}
