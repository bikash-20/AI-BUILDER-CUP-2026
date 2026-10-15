package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestFromContextIncludesRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Timestamp().Logger()

	ctx := WithRequestID(context.Background(), "req-1")
	FromContext(ctx, logger).Info().Msg("hello")

	out := buf.String()
	if !strings.Contains(out, `"request_id":"req-1"`) {
		t.Errorf("missing request id in log: %s", out)
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &parsed); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	if parsed["message"] != "hello" {
		t.Errorf("message = %v, want hello", parsed["message"])
	}
}

func TestRequestIDFromEmpty(t *testing.T) {
	if id := RequestIDFrom(context.Background()); id != "" {
		t.Errorf("expected empty request id, got %q", id)
	}
}
