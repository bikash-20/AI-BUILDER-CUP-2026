package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/finguard/platform/pkg/alloydb"
	"github.com/finguard/platform/pkg/bigquery"
	"github.com/finguard/platform/pkg/domain"
	"github.com/finguard/platform/pkg/httpapi"
	"github.com/finguard/platform/pkg/vertex"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

var validate = validator.New()

// fraudService is the fraud-service application.
type fraudService struct {
	cfg       fraudConfig
	log       zerolog.Logger
	predictor vertex.Predictor
	enricher  bigquery.FeatureClient
	alerts    alloydb.Client
	pub       publisher
	decision  *decisionEngine
}

// fraudConfig defines the runtime knobs fraud-service needs.
type fraudConfig interface {
	ScoreTimeout() time.Duration
}

// publisher is the contract used by fraud-service for emitting events.
type publisher interface {
	PublishJSON(ctx context.Context, topic string, attrs map[string]string, payload any) error
}

// newFraudService wires up the service with its dependencies.
func newFraudService(
	cfg fraudConfig,
	log zerolog.Logger,
	p vertex.Predictor,
	e bigquery.FeatureClient,
	a alloydb.Client,
	pub publisher,
) *fraudService {
	return &fraudService{
		cfg:       cfg,
		log:       log,
		predictor: p,
		enricher:  e,
		alerts:    a,
		pub:       pub,
		decision:  defaultDecisionEngine(),
	}
}

// handleScore implements POST /v1/score.
func (s *fraudService) handleScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpapi.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "POST only")
		return
	}
	var req domain.ScoreRequest
	if err := loadRequest(r, &req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := validate.Struct(&req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "validation_failed", err.Error())
		return
	}

	ctx, cancel := waitDeadline(r, 450*time.Millisecond)
	defer cancel()

	feats, err := s.enricher.GetCustomerFeatures(ctx, req.CustomerID)
	if err != nil {
		s.log.Warn().Err(err).Msg("feature lookup failed; using defaults")
		feats = &bigquery.FeatureRow{CustomerID: req.CustomerID}
	}

	pred, err := s.predictor.Predict(ctx, vertex.Features{
		TxnVelocity1h:  feats.TxnVelocity1h,
		TxnVelocity24h: feats.TxnVelocity24h,
		GeoVelocity:    feats.GeoVelocity,
		DeviceRisk:     feats.DeviceRisk,
		MerchantRisk:   feats.MerchantRisk,
		AmountZScore:   feats.AmountZScore,
	}, req.TxnID)
	if err != nil {
		s.log.Error().Err(err).Msg("predictor failed")
		pred = &vertex.Prediction{FraudProbability: 0.5, ModelVersion: "fallback-v0"}
	}

	decision, reasons := s.decision.Decide(*pred)

	if decision != domain.DecisionApprove {
		_ = s.alerts.InsertAlert(ctx, alloydb.AlertRecord{
			AlertID:    uuid.NewString(),
			TxnID:      req.TxnID,
			CustomerID: req.CustomerID,
			Severity:   severityFromDecision(decision, pred.FraudProbability),
			Reason:     joinReasons(reasons),
			CreatedAt:  time.Now().UTC(),
		})
	}

	_ = s.pub.PublishJSON(ctx, "txn-scored", nil, map[string]any{
		"txn_id":      req.TxnID,
		"customer_id": req.CustomerID,
		"decision":    decision,
		"score":       pred.FraudProbability,
	})

	httpapi.WriteJSON(w, http.StatusOK, map[string]any{
		"decision":          decision,
		"fraud_probability": pred.FraudProbability,
		"risk_score":        int(pred.FraudProbability * 1000),
		"reasons":           reasons,
		"model_version":     pred.ModelVersion,
	})
}

// handleIngest accepts a raw transaction and acknowledges.
func (s *fraudService) handleIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpapi.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "POST only")
		return
	}
	var req domain.Transaction
	if err := loadRequest(r, &req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	httpapi.WriteJSON(w, http.StatusAccepted, map[string]string{"status": "accepted", "txn_id": req.TxnID})
}

// handleListAlerts implements GET /v1/alerts.
//
// Returns a deterministic, simulated list of fraud alerts suitable for the
// analyst console. The data is sourced from alloydb.Client (the same store
// the live service writes to); in mock mode the seed data is returned so
// the UI can render without a database connection.
//
// Query parameters:
//   customer_id  optional, restricts results to a single customer
//   limit        optional, caps the number of rows (default 50, max 200)
func (s *fraudService) handleListAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpapi.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "GET only")
		return
	}

	q := r.URL.Query()
	customerID := q.Get("customer_id")

	limit := 0
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "limit must be a positive integer")
			return
		}
		if n > 200 {
			n = 200
		}
		limit = n
	}

	// Use a short, fixed budget so a slow downstream cannot pin a frontend request.
	ctx, cancel := waitDeadline(r, 1*time.Second)
	defer cancel()

	rows, err := s.alerts.ListAlerts(ctx, alloydb.ListAlertsFilter{
		CustomerID: customerID,
		Limit:      limit,
	})
	if err != nil {
		s.log.Error().Err(err).Msg("list alerts failed")
		httpapi.WriteError(w, http.StatusBadGateway, "upstream_error", err.Error())
		return
	}

	// Map alloydb.AlertRecord -> domain.Alert so the response shape matches
	// what the Next.js BFF already expects for the analyst console. The
	// per-row Simulated flag is derived from the active client: when
	// alloydb.Select returned the MockClient we mark rows as such so the
	// UI can keep showing the simulated-data label honestly.
	simulated := alloydb.IsMock(s.alerts)
	source := "alloydb.MockClient"
	if !simulated {
		source = "alloydb.Postgres"
	}
	out := make([]domain.Alert, 0, len(rows))
	for _, a := range rows {
		out = append(out, domain.Alert{
			AlertID:    a.AlertID,
			TxnID:      a.TxnID,
			CustomerID: a.CustomerID,
			Severity:   domain.Severity(a.Severity),
			Reason:     a.Reason,
			CreatedAt:  a.CreatedAt,
			Status:     domain.AlertStatusOpen,
			Simulated:  simulated,
		})
	}

	httpapi.WriteJSON(w, http.StatusOK, map[string]any{
		"alerts":    out,
		"count":     len(out),
		"simulated": simulated,
		"source":    source,
	})
}

func severityFromDecision(d domain.Decision, p float64) string {
	switch d {
	case domain.DecisionBlock:
		if p > 0.9 {
			return "CRITICAL"
		}
		return "HIGH"
	case domain.DecisionReview:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

func joinReasons(rs []string) string {
	out := ""
	for i, r := range rs {
		if i > 0 {
			out += ","
		}
		out += r
	}
	return out
}
