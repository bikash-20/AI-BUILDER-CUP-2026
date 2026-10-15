package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/finguard/platform/pkg/alloydb"
	"github.com/finguard/platform/pkg/bigquery"
	"github.com/finguard/platform/pkg/domain"
	"github.com/finguard/platform/pkg/pubsub"
	"github.com/finguard/platform/pkg/vertex"
	"github.com/rs/zerolog"
)

type stubConfig struct{ to time.Duration }

func (s stubConfig) ScoreTimeout() time.Duration { return s.to }

type stubPredictor struct{ out vertex.Prediction; err error }

func (s stubPredictor) Predict(_ context.Context, _ vertex.Features, _ string) (*vertex.Prediction, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &s.out, nil
}

func newTestService() *fraudService {
	cfg := stubConfig{to: 200 * time.Millisecond}
	return &fraudService{
		cfg:       &cfg,
		log:       zerolog.New(nil).Level(zerolog.Disabled),
		predictor: stubPredictor{out: vertex.Prediction{FraudProbability: 0.5, ModelVersion: "v1"}},
		enricher:  bigquery.MockClient{},
		alerts:    alloydb.MockClient{},
		pub:       pubsub.NewMockPublisher(),
		decision:  defaultDecisionEngine(),
	}
}

func TestHandleScoreSuccess(t *testing.T) {
	s := newTestService()
	body, _ := json.Marshal(domain.ScoreRequest{
		TxnID: "T-1", CustomerID: "C-1", Amount: 100, Currency: "INR",
		MerchantID: "M-1", Timestamp: time.Now().UTC(),
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/score", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	s.handleScore(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if resp["decision"] != "REVIEW" {
		t.Errorf("decision = %v", resp["decision"])
	}
}

func TestHandleScoreValidation(t *testing.T) {
	s := newTestService()
	req := httptest.NewRequest(http.MethodPost, "/v1/score", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()
	s.handleScore(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestHandleScoreMethodNotAllowed(t *testing.T) {
	s := newTestService()
	req := httptest.NewRequest(http.MethodGet, "/v1/score", nil)
	rec := httptest.NewRecorder()
	s.handleScore(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestHandleScorePredictorFallback(t *testing.T) {
	cfg := stubConfig{to: 200 * time.Millisecond}
	s := &fraudService{
		cfg:       &cfg,
		log:       zerolog.New(nil).Level(zerolog.Disabled),
		predictor: stubPredictor{err: errStub{}},
		enricher:  bigquery.MockClient{},
		alerts:    alloydb.MockClient{},
		pub:       pubsub.NewMockPublisher(),
		decision:  defaultDecisionEngine(),
	}
	body, _ := json.Marshal(domain.ScoreRequest{
		TxnID: "T-2", CustomerID: "C-2", Amount: 100, Currency: "INR",
		MerchantID: "M-1", Timestamp: time.Now().UTC(),
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/score", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	s.handleScore(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 with fallback", rec.Code)
	}
}

func TestSeverityFromDecision(t *testing.T) {
	if got := severityFromDecision(domain.DecisionBlock, 0.95); got != "CRITICAL" {
		t.Errorf("got %q, want CRITICAL", got)
	}
	if got := severityFromDecision(domain.DecisionBlock, 0.7); got != "HIGH" {
		t.Errorf("got %q, want HIGH", got)
	}
	if got := severityFromDecision(domain.DecisionReview, 0.5); got != "MEDIUM" {
		t.Errorf("got %q, want MEDIUM", got)
	}
	if got := severityFromDecision(domain.DecisionApprove, 0.1); got != "LOW" {
		t.Errorf("got %q, want LOW", got)
	}
}

func TestJoinReasons(t *testing.T) {
	if got := joinReasons([]string{"a", "b"}); got != "a,b" {
		t.Errorf("got %q", got)
	}
	if got := joinReasons(nil); got != "" {
		t.Errorf("got %q", got)
	}
}

type errStub struct{}

func (errStub) Error() string { return "stub failure" }

func TestHandleListAlertsSuccess(t *testing.T) {
	s := newTestService()
	req := httptest.NewRequest(http.MethodGet, "/v1/alerts", nil)
	rec := httptest.NewRecorder()
	s.handleListAlerts(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if resp["simulated"] != true {
		t.Errorf("expected simulated=true, got %v", resp["simulated"])
	}
	if resp["source"] != "alloydb.MockClient" {
		t.Errorf("expected source=alloydb.MockClient, got %v", resp["source"])
	}
	alerts, ok := resp["alerts"].([]any)
	if !ok {
		t.Fatalf("alerts missing or wrong type: %T", resp["alerts"])
	}
	if len(alerts) == 0 {
		t.Errorf("expected non-empty list")
	}
	first, ok := alerts[0].(map[string]any)
	if !ok {
		t.Fatalf("first alert not an object: %T", alerts[0])
	}
	if first["simulated"] != true {
		t.Errorf("expected per-row simulated=true, got %v", first["simulated"])
	}
	if first["status"] != "OPEN" {
		t.Errorf("expected status=OPEN, got %v", first["status"])
	}
}

func TestHandleListAlertsCustomerFilter(t *testing.T) {
	s := newTestService()
	req := httptest.NewRequest(http.MethodGet, "/v1/alerts?customer_id=C-1001", nil)
	rec := httptest.NewRecorder()
	s.handleListAlerts(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var resp struct {
		Alerts []domain.Alert `json:"alerts"`
		Count  int            `json:"count"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Count == 0 {
		t.Errorf("expected alerts for C-1001, got 0")
	}
	for _, a := range resp.Alerts {
		if a.CustomerID != "C-1001" {
			t.Errorf("got customer %s, want C-1001", a.CustomerID)
		}
	}
}

func TestHandleListAlertsBadLimit(t *testing.T) {
	s := newTestService()
	req := httptest.NewRequest(http.MethodGet, "/v1/alerts?limit=-1", nil)
	rec := httptest.NewRecorder()
	s.handleListAlerts(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestHandleListAlertsMethodNotAllowed(t *testing.T) {
	s := newTestService()
	req := httptest.NewRequest(http.MethodPost, "/v1/alerts", nil)
	rec := httptest.NewRecorder()
	s.handleListAlerts(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestHandleListAlertsUpstreamFailure(t *testing.T) {
	cfg := stubConfig{to: 200 * time.Millisecond}
	s := &fraudService{
		cfg:      &cfg,
		log:      zerolog.New(nil).Level(zerolog.Disabled),
		alerts:   failingAlertStore{},
		pub:      pubsub.NewMockPublisher(),
		decision: defaultDecisionEngine(),
	}
	req := httptest.NewRequest(http.MethodGet, "/v1/alerts", nil)
	rec := httptest.NewRecorder()
	s.handleListAlerts(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want 502", rec.Code)
	}
}

type failingAlertStore struct{}

func (failingAlertStore) InsertAlert(_ context.Context, _ alloydb.AlertRecord) error { return nil }
func (failingAlertStore) MarkResolved(_ context.Context, _, _ string) error          { return nil }
func (failingAlertStore) ListAlerts(_ context.Context, _ alloydb.ListAlertsFilter) ([]alloydb.AlertRecord, error) {
	return nil, errStub{}
}
