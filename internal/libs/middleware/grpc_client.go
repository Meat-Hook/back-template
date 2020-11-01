package middleware

import (
	"context"
	"net"
	"os"
	"path"

	"github.com/Meat-Hook/back-template/internal/libs/log"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// UnaryClientLogger returns a new unary client interceptor that contains request logger.
func UnaryClientLogger(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	logger := newLogger(ctx, method)
	ctx = logger.WithContext(ctx)

	return invoker(ctx, method, req, reply, cc, opts...)
}

// StreamClientLogger returns a new stream client interceptor that contains request logger.
func StreamClientLogger(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	logger := newLogger(ctx, method)
	ctx = logger.WithContext(ctx)

	return streamer(ctx, desc, cc, method, opts...)
}

// UnaryClientAccessLog returns a new unary client interceptor that logs request status.
func UnaryClientAccessLog(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	err := invoker(ctx, method, req, reply, cc, opts...)
	logger := zerolog.Ctx(ctx)
	logHandler(logger, err)

	return err
}

// StreamClientAccessLog returns a new stream client interceptor that logs request status.
func StreamClientAccessLog(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	logger := zerolog.Ctx(ctx)
	logger.Info().Msg("started")
	clientStream, err := streamer(ctx, desc, cc, method, opts...)
	logHandler(logger, err)

	return clientStream, err
}

func newLogger(ctx context.Context, fullMethod string) zerolog.Logger {
	builder := zerolog.New(os.Stderr).With().Str(log.Func, path.Base(fullMethod))

	if p, ok := peer.FromContext(ctx); ok {
		builder = builder.IPAddr(log.IP, net.ParseIP(p.Addr.String()))
	}

	return builder.Logger()
}

func logHandler(logger *zerolog.Logger, err error) {
	s := status.Convert(err)
	code, msg := s.Code(), s.Message()
	switch code {
	case codes.OK, codes.Canceled, codes.NotFound, codes.AlreadyExists, codes.ResourceExhausted, codes.FailedPrecondition, codes.Aborted:
		logger.Info().Stringer(log.Code, code).Msg(msg)
	case codes.InvalidArgument, codes.DeadlineExceeded, codes.PermissionDenied, codes.OutOfRange, codes.Unavailable, codes.Unauthenticated:
		logger.Warn().Stringer(log.Code, code).Msg(msg)
	default:
		logger.Error().Stringer(log.Code, code).Msg(msg)
	}
}
