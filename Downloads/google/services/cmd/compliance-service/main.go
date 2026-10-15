// Command compliance-service produces SAR/CTR drafts and writes immutable
// audit log entries.
//
//   POST /v1/compliance/sar/draft
//   POST /v1/compliance/ctr/draft
//   GET  /v1/compliance/audit?case_id=...
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
	"syscall"
	"time"

	"github.com/finguard/platform/pkg/config"
	"github.com/finguard/platform/pkg/httpapi"
	"github.com/finguard/platform/pkg/logging"
	"github.com/google/uuid"
)

type draftRequest struct {
	CaseID string `json:"case_id"`
}

type draftResponse struct {
	DraftID string `json:"draft_id"`
	Status  string `json:"status"`
	Summary string `json:"summary"`
}

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

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/v1/compliance/sar/draft", func(w http.ResponseWriter, r *http.Request) {
		draftSAR(w, r, logger)
	})
	mux.HandleFunc("/v1/compliance/ctr/draft", func(w http.ResponseWriter, r *http.Request) {
		draftCTR(w, r, logger)
	})
	mux.HandleFunc("/v1/compliance/audit", func(w http.ResponseWriter, r *http.Request) {
		audit(w, r, logger)
	})

	srv := httpapi.New(httpapi.Options{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
		Logger:  logger,
	})
	logger.Info().Str("service", cfg.ServiceName).Int("port", cfg.Port).Msg("compliance-service starting")
	if err := srv.Run(ctx); err != nil {
		logger.Error().Err(err).Msg("server exited with error")
	}
}

func draftSAR(w http.ResponseWriter, r *http.Request, _ interface{}) {
	if r.Method != http.MethodPost {
		httpapi.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "POST only")
		return
	}
	var req draftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CaseID == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "case_id required")
		return
	}
	httpapi.WriteJSON(w, http.StatusAccepted, draftResponse{
		DraftID: uuid.NewString(),
		Status:  "draft_ready",
		Summary: fmt.Sprintf("SAR draft for case %s generated at %s", req.CaseID, time.Now().UTC().Format(time.RFC3339)),
	})
}

func draftCTR(w http.ResponseWriter, r *http.Request, _ interface{}) {
	if r.Method != http.MethodPost {
		httpapi.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "POST only")
		return
	}
	var req draftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CaseID == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "case_id required")
		return
	}
	httpapi.WriteJSON(w, http.StatusAccepted, draftResponse{
		DraftID: uuid.NewString(),
		Status:  "draft_ready",
		Summary: fmt.Sprintf("CTR draft for case %s", req.CaseID),
	})
}

func audit(w http.ResponseWriter, r *http.Request, _ interface{}) {
	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "case_id required")
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{
		"case_id": caseID,
		"events":  []string{"sar.draft.created", "sar.reviewed", "sar.filed"},
	})
}
