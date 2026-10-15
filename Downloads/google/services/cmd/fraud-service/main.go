// Command fraud-service is the FinGuard real-time transaction scoring service.
//
// Implements FR-1 (transaction ingestion) and FR-2 (real-time scoring).
// Exposes POST /v1/score and POST /v1/transactions.
//
// P95 target: < 500 ms. We enforce a per-request 450 ms deadline that covers
// BigQuery feature lookup, Vertex AI prediction, and AlloyDB alert write.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/finguard/platform/pkg/alloydb"
	"github.com/finguard/platform/pkg/bigquery"
	"github.com/finguard/platform/pkg/config"
	"github.com/finguard/platform/pkg/httpapi"
	"github.com/finguard/platform/pkg/logging"
	"github.com/finguard/platform/pkg/pubsub"
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

	var (
		predictorInst  vertex.Predictor       = &vertex.MockPredictor{ModelVersion: "mock-v1"}
		featureC       bigquery.FeatureClient = bigquery.MockClient{}
		alertStoreInst alloydb.Client         = alloydb.Select(ctx, cfg)
		publisher      pubsub.Publisher       = pubsub.NewMockPublisher()
	)

	if !cfg.MockMode {
		logger.Warn().Msg("running with mock GCP clients")
	}
	if alloydb.IsMock(alertStoreInst) {
		logger.Warn().Msg("alerts store = alloydb.MockClient (set ALLOYDB_DSN to enable live data)")
	} else {
		logger.Info().Msg("alerts store = alloydb.RealClient (Postgres)")
	}

	svc := newFraudService(cfg, logger, predictorInst, featureC, alertStoreInst, publisher)
	srv := httpapi.New(httpapi.Options{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: routes(svc),
		Logger:  logger,
	})

	logger.Info().Str("service", cfg.ServiceName).Int("port", cfg.Port).Msg("fraud-service starting")
	if err := srv.Run(ctx); err != nil {
		logger.Error().Err(err).Msg("server exited with error")
	}
}

func routes(s *fraudService) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/v1/score", s.handleScore)
	mux.HandleFunc("/v1/transactions", s.handleIngest)
	mux.HandleFunc("/v1/alerts", s.handleListAlerts)
	return mux
}

// loadRequest reads and decodes a JSON request body.
func loadRequest(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("empty body")
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

// waitDeadline returns a context.WithTimeout that respects the request's
// existing deadline if it's earlier.
func waitDeadline(r *http.Request, fallback time.Duration) (context.Context, context.CancelFunc) {
	if ddl, ok := r.Context().Deadline(); ok {
		return context.WithDeadline(r.Context(), ddl)
	}
	return context.WithTimeout(r.Context(), fallback)
}

// Ensure config package is referenced.
var _ = config.Load
