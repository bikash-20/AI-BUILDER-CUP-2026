package tracing

import (
	"context"
	"testing"
)

func TestInitNoneExporter(t *testing.T) {
	ctx := context.Background()
	shutdown, err := Init(ctx, "finguard-test", "0.1.0", "none")
	if err != nil {
		t.Fatalf("Init error: %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected non-nil shutdown")
	}
	if err := shutdown(ctx); err != nil {
		t.Errorf("shutdown error: %v", err)
	}
}

func TestTracerReturnsNonNil(t *testing.T) {
	if Tracer("test") == nil {
		t.Fatal("Tracer returned nil")
	}
}
