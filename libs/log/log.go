// Package log stores logged fields, and also provides helper methods for interaction with the logger.
package log

import (
	"context"

	"github.com/rs/zerolog"
)

// Log name.
const (
	Path        = `path`
	GRPCFunc    = `grpc-func`
	HTTPMethod  = `web-method`
	Version     = `version`
	Service     = `service`
	Code        = `code`
	IP          = `ip`
	ReqID       = `req-id`
	User        = `user`
	PanicReason = `panic-reason`
	Duration    = `duration`
	Host        = `host`
	Port        = `port`
	Subsystem   = `subsystem`
	DBMethod    = `db-method`
)

// WarnIfFail logs if callback finished with error.
func WarnIfFail(l zerolog.Logger, cb func() error) {
	if err := cb(); err != nil {
		l.Error().Caller(2).Err(err).Msg("cb fail")
	}
}

type reqIDCtxKey struct{}

// ReqIDWithCtx returns new ctx with request id.
func ReqIDWithCtx(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, reqIDCtxKey{}, reqID)
}

// UnknownID unknown request id.
const UnknownID = "UnknownID"

// ReqIDFromCtx returns request id from context.
// If not found reqID so returns 'UnknownID'.
func ReqIDFromCtx(ctx context.Context) string {
	if reqID, ok := ctx.Value(reqIDCtxKey{}).(string); ok {
		return reqID
	}

	return UnknownID
}
