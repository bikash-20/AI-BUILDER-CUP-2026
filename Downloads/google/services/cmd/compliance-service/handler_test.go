package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDraftSAR(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/compliance/sar/draft",
		strings.NewReader(`{"case_id":"C-1"}`))
	draftSAR(rec, req, nil)
	if rec.Code != http.StatusAccepted {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestDraftSARInvalid(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/compliance/sar/draft", strings.NewReader(`{}`))
	draftSAR(rec, req, nil)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestAudit(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/compliance/audit?case_id=C-1", nil)
	audit(rec, req, nil)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
}
