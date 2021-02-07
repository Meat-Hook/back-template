// Package log stores logged fields, and also provides helper methods for interaction with the logger.
package log

import (
	"github.com/rs/zerolog"
)

// Log name.
const (
	Func          = `func`
	Path          = `path`
	HTTPMethod    = `method`
	Code          = `code`
	IP            = `ip`
	Request       = `request`
	User          = `user`
	Err           = `err`
	HandledStatus = `handled`
	Duration      = `duration`
	Host          = `host`
	Port          = `port`
	Name          = `name`
)

// WarnIfFail logs if callback finished with error.
func WarnIfFail(l zerolog.Logger, cb func() error) {
	if err := cb(); err != nil {
		l.Error().Caller(2).Err(err).Msg("cb fail")
	}
}
