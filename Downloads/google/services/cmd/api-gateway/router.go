package main

import (
	"context"

	"github.com/finguard/platform/pkg/auth"
)

// authMiddlewareNoop is a token-only verifier for tests that bypass real auth.
func authMiddlewareNoop(v auth.Verifier, next func(ctx context.Context)) {
	_, _ = v.Verify(context.Background(), "")
	next(context.Background())
}
