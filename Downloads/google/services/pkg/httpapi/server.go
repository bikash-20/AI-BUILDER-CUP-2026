// Package httpapi contains common HTTP server utilities shared across FinGuard
// Go services: server bootstrap, middleware (request ID, logging, recover),
// and a standard error envelope.
package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/finguard/platform/pkg/logging"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Server wraps http.Server with FinGuard defaults.
type Server struct {
	*http.Server
	logger zerolog.Logger
}

// Options configures a Server.
type Options struct {
	Addr            string
	Handler         http.Handler
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	Logger          zerolog.Logger
}

// New constructs a Server with sensible defaults.
func New(opts Options) *Server {
	if opts.ReadTimeout == 0 {
		opts.ReadTimeout = 5 * time.Second
	}
	if opts.WriteTimeout == 0 {
		opts.WriteTimeout = 30 * time.Second
	}
	if opts.ShutdownTimeout == 0 {
		opts.ShutdownTimeout = 15 * time.Second
	}
	// Wrap with otelhttp + middlewares
	h := Chain(opts.Handler,
		RequestID(),
		Logging(opts.Logger),
		Recover(opts.Logger),
	)
	h = otelhttp.NewHandler(h, "finguard")

	return &Server{
		Server: &http.Server{
			Addr:         opts.Addr,
			Handler:      h,
			ReadTimeout:  opts.ReadTimeout,
			WriteTimeout: opts.WriteTimeout,
		},
		logger: opts.Logger,
	}
}

// Run starts the server and blocks until ctx is cancelled.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info().Str("addr", s.Addr).Msg("http server starting")
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return s.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

// Logging returns an HTTP middleware that logs requests using zerolog.
func Logging(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)
			logger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", rw.status).
				Dur("dur", time.Since(start)).
				Str("request_id", logging.RequestIDFrom(r.Context())).
				Msg("http")
		})
	}
}

// RequestID sets X-Request-Id on the request context and response header.
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-Id")
			if id == "" {
				id = uuid.NewString()
			}
			w.Header().Set("X-Request-Id", id)
			ctx := logging.WithRequestID(r.Context(), id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Recover converts panics into a 500 response with the standard error envelope.
func Recover(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rv := recover(); rv != nil {
					logger.Error().Interface("panic", rv).Str("path", r.URL.Path).Msg("panic")
					WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Chain applies middleware in order.
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}
