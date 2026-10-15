package alloydb

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// These tests exercise the live Postgres path against a real Postgres
// reachable via TEST_ALLOYDB_DSN. They are skipped when the env var is
// unset so the standard `go test ./...` run stays offline.
//
// To run them locally:
//
//   docker compose up -d postgres
//   bash scripts/db_bootstrap.sh
//   TEST_ALLOYDB_DSN=postgres://finguard:finguard@127.0.0.1:5433/finguard?sslmode=disable \
//     go test ./pkg/alloydb/...

const testAlertsSchema = `
DROP TABLE IF EXISTS alerts;
CREATE TABLE alerts (
  alert_id      TEXT PRIMARY KEY,
  txn_id        TEXT NOT NULL,
  customer_id   TEXT NOT NULL,
  severity      TEXT NOT NULL,
  reason        TEXT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  status        TEXT NOT NULL DEFAULT 'OPEN',
  closure_note  TEXT,
  closed_at     TIMESTAMPTZ,
  model_version TEXT
);
`

func openTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_ALLOYDB_DSN")
	if dsn == "" {
		t.Skip("TEST_ALLOYDB_DSN not set; skipping live alloydb test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	if _, err := pool.Exec(ctx, testAlertsSchema); err != nil {
		pool.Close()
		t.Fatalf("reset schema: %v", err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}

func TestRealClientListAlerts(t *testing.T) {
	pool := openTestPool(t)
	c := &RealClient{Pool: pool}
	ctx := context.Background()

	now := time.Date(2026, 10, 15, 10, 0, 0, 0, time.UTC)
	want := []AlertRecord{
		{AlertID: "ALT-live-1", TxnID: "T-1", CustomerID: "C-1001", Severity: "HIGH", Reason: "device_risk", CreatedAt: now},
		{AlertID: "ALT-live-2", TxnID: "T-2", CustomerID: "C-1002", Severity: "LOW", Reason: "txn_velocity_1h", CreatedAt: now.Add(-time.Hour)},
	}
	for _, a := range want {
		if err := c.InsertAlert(ctx, a); err != nil {
			t.Fatalf("insert %s: %v", a.AlertID, err)
		}
	}
	// MarkResolved test
	if err := c.MarkResolved(ctx, "ALT-live-1", "ok"); err != nil {
		t.Fatalf("mark resolved: %v", err)
	}

	got, err := c.ListAlerts(ctx, ListAlertsFilter{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 rows, got %d: %+v", len(got), got)
	}
	if got[0].AlertID != "ALT-live-1" {
		t.Errorf("expected most-recent first; got %s", got[0].AlertID)
	}

	gotByCust, err := c.ListAlerts(ctx, ListAlertsFilter{CustomerID: "C-1002"})
	if err != nil {
		t.Fatalf("list by customer: %v", err)
	}
	if len(gotByCust) != 1 || gotByCust[0].AlertID != "ALT-live-2" {
		t.Errorf("customer filter wrong: %+v", gotByCust)
	}

	gotLimit, err := c.ListAlerts(ctx, ListAlertsFilter{Limit: 1})
	if err != nil {
		t.Fatalf("list limit: %v", err)
	}
	if len(gotLimit) != 1 {
		t.Errorf("expected 1 row with limit=1, got %d", len(gotLimit))
	}
}

func TestSelectReal(t *testing.T) {
	dsn := os.Getenv("TEST_ALLOYDB_DSN")
	if dsn == "" {
		t.Skip("TEST_ALLOYDB_DSN not set; skipping Select test")
	}
	t.Setenv("ALLOYDB_DSN", dsn)
	c, err := NewRealClient(context.Background(), dsn)
	if err != nil {
		t.Fatalf("new real client: %v", err)
	}
	if IsMock(c) {
		t.Error("expected non-mock client when TEST_ALLOYDB_DSN is set")
	}
}

func TestSelectMockFallback(t *testing.T) {
	// No env var -> mock path.
	t.Setenv("ALLOYDB_DSN", "")
	c := Select(context.Background(), nil)
	if !IsMock(c) {
		t.Errorf("expected mock client, got %T", c)
	}
}