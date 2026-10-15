// Command db-bootstrap applies the local-dev Postgres schema for FinGuard
// and seeds the same five alerts the in-process MockClient returns today.
//
// It is idempotent: schema uses IF NOT EXISTS; alert rows are inserted
// with ON CONFLICT (alert_id) DO NOTHING.
//
// Required env:
//   DATABASE_URL — pgx-style DSN, e.g.
//     postgres://finguard:finguard@127.0.0.1:5433/finguard?sslmode=disable
//
// The script exits 0 on success, non-zero with a clear message otherwise.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// alertsSchema is the Postgres port of the production AlloyDB alerts table.
// Kept inline so this binary has no extra file dependencies.
const alertsSchema = `
CREATE TABLE IF NOT EXISTS alerts (
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
CREATE INDEX IF NOT EXISTS alerts_customer_created_idx ON alerts (customer_id, created_at DESC);
CREATE INDEX IF NOT EXISTS alerts_created_idx ON alerts (created_at DESC);
`

// seedAlert mirrors alloydb.seedAlerts so a freshly-bootstrapped DB
// returns the same data the MockClient does today.
type seedAlert struct {
	id, txn, cust, sev, reason string
	createdAt                  time.Time
}

var seedAlerts = []seedAlert{
	{"ALT-seed-001", "T-1003", "C-1001", "HIGH", "geo_velocity,device_risk", time.Date(2026, 10, 15, 9, 12, 0, 0, time.UTC)},
	{"ALT-seed-002", "T-1002", "C-1001", "MEDIUM", "amount_z_score", time.Date(2026, 10, 15, 8, 47, 0, 0, time.UTC)},
	{"ALT-seed-003", "T-0998", "C-1002", "CRITICAL", "merchant_risk,amount_z_score", time.Date(2026, 10, 15, 7, 30, 0, 0, time.UTC)},
	{"ALT-seed-004", "T-0995", "C-1003", "LOW", "txn_velocity_1h", time.Date(2026, 10, 15, 6, 5, 0, 0, time.UTC)},
	{"ALT-seed-005", "T-0990", "C-1002", "HIGH", "device_risk,merchant_risk", time.Date(2026, 10, 14, 22, 14, 0, 0, time.UTC)},
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required (e.g. postgres://finguard:finguard@127.0.0.1:5433/finguard?sslmode=disable)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	if _, err := pool.Exec(ctx, alertsSchema); err != nil {
		log.Fatalf("apply schema: %v", err)
	}
	fmt.Println("schema: alerts table ready")

	const insertQ = `INSERT INTO alerts (alert_id, txn_id, customer_id, severity, reason, created_at, status, model_version)
	                 VALUES ($1, $2, $3, $4, $5, $6, 'OPEN', 'bootstrap-v1')
	                 ON CONFLICT (alert_id) DO NOTHING`
	inserted := 0
	for _, a := range seedAlerts {
		ct, err := pool.Exec(ctx, insertQ, a.id, a.txn, a.cust, a.sev, a.reason, a.createdAt)
		if err != nil {
			log.Fatalf("insert %s: %v", a.id, err)
		}
		inserted += int(ct.RowsAffected())
	}
	fmt.Printf("alerts: seeded %d new rows (out of %d total)\n", inserted, len(seedAlerts))
}
