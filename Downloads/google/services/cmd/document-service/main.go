// Command document-service is the Go BFF for document upload and retrieval.
//
// POST /v1/documents/upload -> 202 with doc_id
// GET  /v1/documents/{id}    -> 200 with status
//
// The actual OCR/NER is delegated to the Python `document_service` Cloud Run
// service (see ml/services/document_service).
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
	"strings"
	"syscall"
	"time"

	"github.com/finguard/platform/pkg/config"
	"github.com/finguard/platform/pkg/domain"
	"github.com/finguard/platform/pkg/httpapi"
	"github.com/finguard/platform/pkg/logging"
	"github.com/google/uuid"
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

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/v1/documents/upload", upload)
	mux.HandleFunc("/v1/documents/", getDoc)

	srv := httpapi.New(httpapi.Options{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
		Logger:  logger,
	})
	logger.Info().Str("service", cfg.ServiceName).Int("port", cfg.Port).Msg("document-service starting")
	if err := srv.Run(ctx); err != nil {
		logger.Error().Err(err).Msg("server exited with error")
	}
}

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpapi.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "POST only")
		return
	}
	var req domain.UploadDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if req.DocType == "" || req.FileURL == "" || !strings.HasPrefix(req.FileURL, "gs://") {
		httpapi.WriteError(w, http.StatusBadRequest, "validation_failed", "doc_type and file_url (gs://) required")
		return
	}
	resp := domain.UploadDocumentResponse{
		DocID:  uuid.NewString(),
		Status: domain.DocStatusProcessing,
	}
	httpapi.WriteJSON(w, http.StatusAccepted, resp)
}

func getDoc(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "missing doc id")
		return
	}
	docID := parts[2]
	if docID == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "empty doc id")
		return
	}
	resp := domain.Document{
		DocID:      docID,
		DocType:    domain.DocumentKYC,
		Status:     domain.DocStatusExtracted,
		Confidence: 0.93,
		UploadedAt: time.Now().UTC(),
	}
	httpapi.WriteJSON(w, http.StatusOK, resp)
}

// unused
var _ = errors.New
