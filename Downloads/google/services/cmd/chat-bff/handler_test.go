package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/finguard/platform/pkg/dlp"
)

func TestHandleChatSuccess(t *testing.T) {
	red := dlp.NewRedactor()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat",
		strings.NewReader(`{"message":"hello"}`))
	handleChat(rec, req, nil, red)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var resp chatResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.SessionID == "" {
		t.Error("expected session_id to be generated")
	}
}

func TestHandleChatEmpty(t *testing.T) {
	red := dlp.NewRedactor()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat", strings.NewReader(`{}`))
	handleChat(rec, req, nil, red)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestHandleChatPIIRedaction(t *testing.T) {
	red := dlp.NewRedactor()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat",
		strings.NewReader(`{"message":"contact alice@example.com"}`))
	handleChat(rec, req, nil, red)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "[REDACTED_EMAIL]") {
		t.Errorf("expected redaction, got %s", rec.Body.String())
	}
}
