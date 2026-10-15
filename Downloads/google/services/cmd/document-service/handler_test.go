package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUploadSuccess(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/documents/upload",
		strings.NewReader(`{"doc_type":"KYC","file_url":"gs://x/y.pdf","customer_id":"C-1"}`))
	upload(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Errorf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestUploadInvalidBody(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/documents/upload", strings.NewReader(`{}`))
	upload(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestUploadMethodNotAllowed(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/documents/upload", nil)
	upload(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestGetDoc(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/documents/D-1", nil)
	getDoc(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
}
