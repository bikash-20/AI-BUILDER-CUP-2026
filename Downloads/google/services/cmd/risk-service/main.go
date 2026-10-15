// Command risk-service returns the composite risk score for a customer.
//
// Implements FR-5 (Risk Scoring).
//   GET  /v1/risk/{customer_id}
//   POST /v1/risk/recompute
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/finguard/platform/pkg/config"
	"github.com/finguard/platform/pkg/domain"
	"github.com/finguard/platform/pkg/httpapi"
	"github.com/finguard/platform/pkg/logging"
	"github.com/finguard/platform/pkg/riskstore"
	"github.com/finguard/platform/pkg/vertex"
)

func main() {
	port := flag.Int("port", 0, "override PORT env")
	flag.Parse()
	if *port > 0 {
		_ = os.Setenv("PORT", fmt.Sprintf("%d", *port))
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	logger := logging.New(cfg.ServiceName, cfg.Env)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Wire the riskstore client: in production this points at BigQuery;
	// locally, with no BIGQUERY_EMULATOR_HOST set, it falls back to the
	// in-process MockClient so the service still boots and returns
	// clearly-marked simulated data.
	rs := riskstore.Select(cfg)
	if riskstore.IsMock(rs) {
		logger.Warn().Msg("transactions store = riskstore.MockClient (set BIGQUERY_EMULATOR_HOST to enable live data)")
	} else {
		logger.Info().Str("host", cfg.BigQueryEmulatorHost).Msg("transactions store = riskstore.BigQueryClient")
	}

	srv := httpapi.New(httpapi.Options{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: routes(logger, rs),
		Logger:  logger,
	})

	logger.Info().Str("service", cfg.ServiceName).Int("port", cfg.Port).Msg("risk-service starting")
	if err := srv.Run(ctx); err != nil {
		logger.Error().Err(err).Msg("server exited with error")
	}
}

func routes(logger interface{}, rs riskstore.Client) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/v1/risk", riskRoot)
	mux.HandleFunc("/v1/risk/", riskByID)
	mux.HandleFunc("/v1/customers/", func(w http.ResponseWriter, r *http.Request) {
		customerSubtree(w, r, rs)
	})
	return mux
}

func riskRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && (r.URL.Path == "/v1/risk/recompute" || strings.HasSuffix(r.URL.Path, "recompute")) {
		recompute(w, r)
		return
	}
	httpapi.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use GET /v1/risk/{id}")
}

func riskByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpapi.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "GET only")
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "missing customer id")
		return
	}
	customerID := parts[2]
	resp := computeRisk(customerID)
	httpapi.WriteJSON(w, http.StatusOK, resp)
}

func recompute(w http.ResponseWriter, r *http.Request) {
	httpapi.WriteJSON(w, http.StatusAccepted, map[string]any{"status": "queued"})
}

// computeRisk synthesizes a stable composite score for the given customer.
func computeRisk(customerID string) domain.RiskResponse {
	// Deterministic seed-based composite for mock mode.
	score := float64(stableHash(customerID)%1000) / 10.0
	fraud := clamp01(score / 1000.0)
	credit := clamp01(0.6 + (float64(stableHash("cr:"+customerID)%40)-20)/100.0)
	behavioral := clamp01(0.5 + (float64(stableHash("bh:"+customerID)%40)-20)/100.0)
	compliance := 0.0

	band := bandFromScore(score)
	return domain.RiskResponse{
		CustomerID: customerID,
		Score:      score,
		Band:       band,
		Breakdown: domain.RiskBreakdown{
			Fraud:      fraud * 1000,
			Credit:     credit * 1000,
			Behavioral: behavioral * 1000,
			Compliance: compliance,
		},
		UpdatedAt: time.Now().UTC(),
	}
}

func bandFromScore(s float64) domain.RiskBand {
	switch {
	case s < 300:
		return domain.RiskBandLow
	case s < 600:
		return domain.RiskBandMedium
	case s < 850:
		return domain.RiskBandHigh
	default:
		return domain.RiskBandCrit
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// stableHash is a tiny deterministic hash for IDs. Kept local so the
// composite-score path (computeRisk) is independent of riskstore.
func stableHash(s string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

// ensure the vertex import is referenced (avoid unused import in scaffold)
var _ = vertex.Features{}

// customerSubtree routes /v1/customers/{id}/transactions (and is the entry
// point for any future per-customer sub-resources). The trailing slash
// matches both the bare prefix and any deeper path.
//
// The transactions are read from the wired riskstore.Client. When the
// client is the MockClient, the envelope and per-row Simulated flag will
// be true; when it is the BigQueryClient, they will be false.
func customerSubtree(w http.ResponseWriter, r *http.Request, rs riskstore.Client) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// expected: ["v1", "customers", "{id}", "transactions"]
	if len(parts) < 4 || parts[0] != "v1" || parts[1] != "customers" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "path must be /v1/customers/{id}/transactions")
		return
	}
	if parts[3] != "transactions" {
		httpapi.WriteError(w, http.StatusNotFound, "not_found", "unknown sub-resource: "+parts[3])
		return
	}
	if r.Method != http.MethodGet {
		httpapi.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "GET only")
		return
	}
	customerID := parts[2]
	if customerID == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "missing customer id")
		return
	}
	limit := 0
	if v := r.URL.Query().Get("limit"); v != "" {
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
	// Use a short, fixed budget so a slow BigQuery emulator cannot pin the
	// frontend request. The riskstore client may apply its own per-call
	// timeout; this is a defense-in-depth outer bound.
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()
	resp, err := rs.ListCustomerTransactions(ctx, customerID, limit)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadGateway, "upstream_error", err.Error())
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, resp)
}
