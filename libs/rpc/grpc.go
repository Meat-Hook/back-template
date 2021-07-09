// Package rpc provide helpers for typical gRPC client/server.
package rpc

import (
	"context"
	"net"
	"path"
	"time"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/Meat-Hook/back-template/libs/log"
	"github.com/Meat-Hook/back-template/libs/metrics"
)

const (
	keepaliveTime    = 50 * time.Second
	keepaliveTimeout = 10 * time.Second
	keepaliveMinTime = 30 * time.Second
)

var (
	errInternal = status.Error(codes.Internal, "internal error")
)

func newRPCLogger(ctx context.Context, logger zerolog.Logger, fullMethod string) zerolog.Logger {
	l := logger.With().
		Str(log.Func, path.Base(fullMethod)).
		Logger()

	if p, ok := peer.FromContext(ctx); ok {
		host, _, err := net.SplitHostPort(p.Addr.String())
		if err != nil {
			l.Error().Err(err).Msg("net: split host and port")
		} else {
			l = l.With().IPAddr(log.IP, net.ParseIP(host)).Logger()
		}
	}

	return l
}

func rpcLogHandler(l *zerolog.Logger, err error) error {
	s := status.Convert(err)

	code, msg := s.Code(), s.Message()
	switch code {
	case codes.OK, codes.Canceled, codes.NotFound:
		l.Info().Str(log.Code, code.String()).Str(log.HandledStatus, "success").Send()
	case codes.Unknown:
		l.Error().Str(log.Code, code.String()).Str(log.HandledStatus, "failed").Msg(msg)
		err = errInternal
	default:
		l.Error().Str(log.Code, code.String()).Str(log.HandledStatus, "failed").Msg(msg)
	}

	return err
}

var _ grpc_recovery.RecoveryHandlerFuncContext = recoveryFunc

func recoveryFunc(ctx context.Context, p interface{}) error {
	metrics.PanicsTotal.Inc()
	l := zerolog.Ctx(ctx)
	l.Error().Stack().
		Uint32(log.Code, uint32(codes.Internal)).
		Interface(log.PanicReason, p).Stack().Msg("panic")

	return errInternal
}
