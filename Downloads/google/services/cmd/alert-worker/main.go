// Command alert-worker is a background process that consumes alerts from
// Pub/Sub and pushes them to the analyst console via WebSocket fanout.
//
// In mock mode it polls the in-memory publisher.
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
	"sync"
	"syscall"
	"time"

	"github.com/finguard/platform/pkg/config"
	"github.com/finguard/platform/pkg/logging"
	"github.com/gorilla/websocket"
)

type alertMsg struct {
	AlertID    string  `json:"alert_id"`
	TxnID      string  `json:"txn_id"`
	CustomerID string  `json:"customer_id"`
	Severity   string  `json:"severity"`
	Score      float64 `json:"score"`
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

type hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

func newHub() *hub { return &hub{clients: map[*websocket.Conn]struct{}{}} }

func (h *hub) add(c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = struct{}{}
}
func (h *hub) remove(c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, c)
	_ = c.Close()
}
func (h *hub) broadcast(msg any) {
	h.mu.Lock()
	defer h.mu.Unlock()
	data, _ := json.Marshal(msg)
	for c := range h.clients {
		_ = c.WriteMessage(websocket.TextMessage, data)
	}
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

	h := newHub()
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/ws/alerts", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Warn().Err(err).Msg("ws upgrade")
			return
		}
		h.add(c)
		logger.Info().Str("remote", r.RemoteAddr).Msg("analyst connected")
		// Read loop just to detect close.
		for {
			if _, _, err := c.ReadMessage(); err != nil {
				h.remove(c)
				return
			}
		}
	})

	// In mock mode, generate synthetic alerts every 5s.
	if cfg.MockMode {
		go func() {
			t := time.NewTicker(5 * time.Second)
			defer t.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-t.C:
					h.broadcast(alertMsg{
						AlertID:    fmt.Sprintf("A-%d", time.Now().Unix()),
						TxnID:      "T-mock",
						CustomerID: "C-mock",
						Severity:   "MEDIUM",
						Score:      0.5,
					})
				}
			}
		}()
	}

	srv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.Port), Handler: mux}
	logger.Info().Str("service", cfg.ServiceName).Int("port", cfg.Port).Msg("alert-worker starting")
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("server error")
		}
	}()

	<-ctx.Done()
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutCtx)
}
