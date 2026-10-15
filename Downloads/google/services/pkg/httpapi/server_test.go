package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

func newTestLogger() zerolog.Logger { return zerolog.New(nil).Level(zerolog.Disabled) }

func TestRequestIDPropagates(t *testing.T) {
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`"ok"`))
	}), RequestID())

	req := httptest.NewRequest("GET", "/x", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-Id"); got == "" {
		t.Error("expected request id header")
	}
}

func TestRecoverPanic(t *testing.T) {
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}), Recover(newTestLogger()), ContentTypeJSON)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/x", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d", rec.Code)
	}
	var env ErrorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if env.Error.Code != "internal_error" {
		t.Errorf("code = %q", env.Error.Code)
	}
}

func TestMaxBodyBytesRejects(t *testing.T) {
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), MaxBodyBytes(10))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/x", nil)
	req.ContentLength = 100
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d", rec.Code)
	}
}
