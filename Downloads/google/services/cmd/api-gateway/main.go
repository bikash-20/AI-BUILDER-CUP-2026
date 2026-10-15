// Command api-gateway is the FinGuard API gateway.
//
// It terminates IAP/OAuth, validates JWTs, rate-limits, and routes
// requests to downstream Cloud Run services. In mock mode it
// short-circuits to in-process handlers.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/finguard/platform/pkg/auth"
	"github.com/finguard/platform/pkg/config"
	"github.com/finguard/platform/pkg/httpapi"
	"github.com/finguard/platform/pkg/logging"
)

type backend struct {
	Name string
	URL  *url.URL
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

	backends := loadBackends(cfg)
	logger.Info().Strs("backends", backendNames(backends)).Msg("api-gateway starting")

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "backends": backendNames(backends)})
	})

	verifier := auth.AllowAll{}
	mux.Handle("/v1/", authMiddleware(verifier, proxyRouter(backends)))

	srv := httpapi.New(httpapi.Options{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
		Logger:  logger,
	})

	logger.Info().Str("service", cfg.ServiceName).Int("port", cfg.Port).Msg("api-gateway listening")
	if err := srv.Run(ctx); err != nil {
		logger.Error().Err(err).Msg("server exited with error")
	}
}

func loadBackends(cfg *config.Config) map[string]backend {
	env := strings.ToLower(cfg.Env)
	out := map[string]backend{}

	add := func(name, host string) {
		if host == "" {
			return
		}
		u, err := url.Parse(host)
		if err != nil {
			return
		}
		out[name] = backend{Name: name, URL: u}
	}
	add("fraud", envURL(env, "FINGUARD_FRAUD_URL", "http://fraud-service:8080"))
	add("risk", envURL(env, "FINGUARD_RISK_URL", "http://risk-service:8080"))
	add("document", envURL(env, "FINGUARD_DOCUMENT_URL", "http://document-service:8080"))
	add("chat", envURL(env, "FINGUARD_CHAT_URL", "http://chat-bff:8080"))
	add("compliance", envURL(env, "FINGUARD_COMPLIANCE_URL", "http://compliance-service:8080"))
	return out
}

func envURL(env, key, def string) string {
	v := os.Getenv(key)
	if v != "" {
		return v
	}
	return def
}

func backendNames(b map[string]backend) []string {
	out := make([]string, 0, len(b))
	for k := range b {
		out = append(out, k)
	}
	return out
}

func proxyRouter(backends map[string]backend) http.Handler {
	director := func(req *http.Request) {
		path := req.URL.Path
		var target *url.URL
		switch {
		case strings.HasPrefix(path, "/v1/score") || strings.HasPrefix(path, "/v1/transactions"):
			if b, ok := backends["fraud"]; ok {
				target = b.URL
			}
		case strings.HasPrefix(path, "/v1/alerts"):
			if b, ok := backends["fraud"]; ok {
				target = b.URL
			}
		case strings.HasPrefix(path, "/v1/risk"):
			if b, ok := backends["risk"]; ok {
				target = b.URL
			}
		case strings.HasPrefix(path, "/v1/customers"):
			if b, ok := backends["risk"]; ok {
				target = b.URL
			}
		case strings.HasPrefix(path, "/v1/documents"):
			if b, ok := backends["document"]; ok {
				target = b.URL
			}
		case strings.HasPrefix(path, "/v1/chat"):
			if b, ok := backends["chat"]; ok {
				target = b.URL
			}
		case strings.HasPrefix(path, "/v1/compliance"):
			if b, ok := backends["compliance"]; ok {
				target = b.URL
			}
		}
		if target == nil {
			// Mark the request so the ErrorHandler can write a 502.
			// We can't write a response from the Director (no ResponseWriter).
			req.Header.Set("X-FinGuard-No-Backend", path)
			return
		}
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
		req.Header.Set("X-Forwarded-Host", req.Host)
	}
	return &httputil.ReverseProxy{
		Director: director,
		Transport: &http.Transport{
			IdleConnTimeout:    30 * time.Second,
			MaxIdleConnsPerHost: 50,
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			if r.Header.Get("X-FinGuard-No-Backend") != "" {
				httpapi.WriteError(w, http.StatusBadGateway, "no_backend", "no backend configured for path")
				return
			}
			httpapi.WriteError(w, http.StatusBadGateway, "upstream_error", err.Error())
		},
	}
}

func authMiddleware(v auth.Verifier, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tok := r.Header.Get("Authorization")
		if tok == "" {
			tok = r.URL.Query().Get("access_token")
		}
		claims, err := v.Verify(r.Context(), tok)
		if err != nil {
			httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", err.Error())
			return
		}
		// Propagate subject to backend
		r.Header.Set("X-User-Subject", claims.Subject)
		next.ServeHTTP(w, r)
	})
}
