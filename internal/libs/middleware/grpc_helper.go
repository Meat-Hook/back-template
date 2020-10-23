package middleware

import (
	"context"
	"net"
	"path"

	"github.com/Meat-Hook/back-template/internal/libs/log"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func RPCLogHandler(l *zerolog.Logger, err error) {
	s := status.Convert(err)

	code, msg := s.Code(), s.Message()
	switch code {
	case codes.OK, codes.Canceled, codes.NotFound:
		l.Info().Str(log.Code, code.String()).Str(log.HandledStatus, "success").Send()
	default:
		l.Error().Str(log.Code, code.String()).Str(log.HandledStatus, "failed").Msg(msg)
	}
}

func newRPCLogger(ctx context.Context, logBuilder zerolog.Context, fullMethod string) zerolog.Logger {
	reqID := xid.New()

	l := logBuilder.
		Str(log.Func, path.Base(fullMethod)).
		Str(log.Request, reqID.String()).
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
