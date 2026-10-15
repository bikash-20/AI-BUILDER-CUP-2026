// Package logging provides a structured JSON logger compatible with
// Google Cloud Logging.
//
// The Logger emits one structured JSON object per line on stdout/stderr,
// which Cloud Run's log agent ships to Cloud Logging without further processing.
package logging

import (
	"context"
	"os"

	"github.com/rs/zerolog"
)

type ctxKey int

const (
	requestIDKey ctxKey = iota
)

// New creates a JSON logger with the given service name and environment.
func New(service, env string) zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.MessageFieldName = "message"
	zerolog.LevelFieldName = "severity"

	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", service).
		Str("env", env).
		Logger()
	return logger
}

// WithRequestID stores a request ID on the context.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestIDFrom returns the request ID stored on the context, or "".
func RequestIDFrom(ctx context.Context) string {
	v, _ := ctx.Value(requestIDKey).(string)
	return v
}

// FromContext returns a child logger pre-populated with the request ID.
func FromContext(ctx context.Context, base zerolog.Logger) *zerolog.Logger {
	if id := RequestIDFrom(ctx); id != "" {
		l := base.With().Str("request_id", id).Logger()
		return &l
	}
	return &base
}
