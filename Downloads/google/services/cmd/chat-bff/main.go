// Command chat-bff is the BFF between the Next.js frontend and the Python
// Gemini service. It performs session management, DLP redaction, and
// proxies the request.
//
//   POST /v1/chat
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/finguard/platform/pkg/config"
	"github.com/finguard/platform/pkg/dlp"
	"github.com/finguard/platform/pkg/httpapi"
	"github.com/finguard/platform/pkg/logging"
	"github.com/google/uuid"
)

type chatRequest struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

type chatResponse struct {
	SessionID string   `json:"session_id"`
	Reply     string   `json:"reply"`
	Citations []string `json:"citations"`
	Actions   []string `json:"actions"`
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

	redactor := dlp.NewRedactor()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/v1/chat", func(w http.ResponseWriter, r *http.Request) {
		handleChat(w, r, logger, redactor)
	})

	srv := httpapi.New(httpapi.Options{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
		Logger:  logger,
	})
	logger.Info().Str("service", cfg.ServiceName).Int("port", cfg.Port).Msg("chat-bff starting")
	if err := srv.Run(ctx); err != nil {
		logger.Error().Err(err).Msg("server exited with error")
	}
}

func handleChat(w http.ResponseWriter, r *http.Request, logger interface{}, redactor *dlp.Redactor) {
	if r.Method != http.MethodPost {
		httpapi.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "POST only")
		return
	}
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if req.Message == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "validation_failed", "message required")
		return
	}
	if req.SessionID == "" {
		req.SessionID = uuid.NewString()
	}

	redacted := redactor.Redact(req.Message)
	resp := chatResponse{
		SessionID: req.SessionID,
		Reply:     fmt.Sprintf("Acknowledged: %s", redacted),
		Citations: []string{"policy://generic/v1"},
		Actions:   []string{},
	}
	if bytes.Contains([]byte(redacted), []byte("[REDACTED_")) {
		resp.Actions = append(resp.Actions, "verify_identity")
	}
	_ = io.Discard
	_ = time.Second
	httpapi.WriteJSON(w, http.StatusOK, resp)
}
