package middleware

import (
	"context"
	"net"
	"path"
	"strings"

	"github.com/Meat-Hook/back-template/internal/libs/log"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func rpcLogHandler(l *zerolog.Logger, err error) {
	s := status.Convert(err)

	code, msg := s.Code(), s.Message()
	switch code {
	case codes.OK, codes.Canceled, codes.NotFound:
		l.Info().Str(log.Code, code.String()).Str(log.HandledStatus, "success").Send()
	default:
		l.Error().Str(log.Code, code.String()).Str(log.HandledStatus, "failed").Msg(msg)
	}
}

func newRPCLogger(ctx context.Context, logger zerolog.Logger, md metadata.MD, fullMethod string) (zerolog.Logger, string) {
	reqID := log.UnknownID
	if res := md.Get(log.ReqID); res != nil {
		reqID = strings.Join(res, "")
	}

	l := logger.With().
		Str(log.Func, path.Base(fullMethod)).
		Str(log.ReqID, reqID).
		Logger()

	if p, ok := peer.FromContext(ctx); ok {
		host, _, err := net.SplitHostPort(p.Addr.String())
		if err != nil {
			l.Error().Err(err).Msg("net: split host and port")
		} else {
			l = l.With().IPAddr(log.IP, net.ParseIP(host)).Logger()
		}
	}

	return l, reqID
}
