package alloydb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/finguard/platform/pkg/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RealClient is the Postgres-backed implementation of Client. It uses
// the pgx connection pool and is the local-dev proxy for the production
// AlloyDB cluster (AlloyDB is wire-compatible with Postgres).
//
// Returned records carry no "simulated" marker; downstream code uses
// alloydb.IsMock to set the per-row and envelope Simulated flags.
type RealClient struct {
	Pool *pgxpool.Pool
}

// NewRealClient opens a pgx pool against the given DSN. Callers are
// responsible for closing the pool via Pool.Close() during shutdown.
func NewRealClient(ctx context.Context, dsn string) (*RealClient, error) {
	if dsn == "" {
		return nil, errors.New("alloydb: empty dsn")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("alloydb: parse dsn: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnIdleTime = 5 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("alloydb: connect: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("alloydb: ping: %w", err)
	}
	return &RealClient{Pool: pool}, nil
}

// InsertAlert writes a single alert to the alerts table.
func (r *RealClient) InsertAlert(ctx context.Context, a AlertRecord) error {
	if a.AlertID == "" {
		return errors.New("alloydb: empty alert id")
	}
	const q = `INSERT INTO alerts (alert_id, txn_id, customer_id, severity, reason, created_at, status, model_version)
	           VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	           ON CONFLICT (alert_id) DO NOTHING`
	_, err := r.Pool.Exec(ctx, q,
		a.AlertID, a.TxnID, a.CustomerID, a.Severity, a.Reason, a.CreatedAt, "OPEN", a.ModelVersion,
	)
	return err
}

// MarkResolved updates the alerts row to CLOSED with the given resolution.
func (r *RealClient) MarkResolved(ctx context.Context, alertID, resolution string) error {
	if alertID == "" {
		return errors.New("alloydb: empty alert id")
	}
	const q = `UPDATE alerts SET status = 'CLOSED', closure_note = $2, closed_at = now() WHERE alert_id = $1`
	_, err := r.Pool.Exec(ctx, q, alertID, resolution)
	return err
}

// ListAlerts returns the most-recent alerts, optionally filtered by
// customer id and capped by f.Limit. Rows are returned newest-first.
func (r *RealClient) ListAlerts(ctx context.Context, f ListAlertsFilter) ([]AlertRecord, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = defaultAlertLimit
	}
	if limit > 200 {
		limit = 200
	}
	const base = `SELECT alert_id, txn_id, customer_id, severity, reason, created_at
	              FROM alerts`
	var (
		rows interface {
			Close()
			Next() bool
			Scan(dest ...any) error
			Err() error
		}
		err error
	)
	if f.CustomerID != "" {
		rows, err = r.Pool.Query(ctx, base+" WHERE customer_id = $1 ORDER BY created_at DESC LIMIT $2", f.CustomerID, limit)
	} else {
		rows, err = r.Pool.Query(ctx, base+" ORDER BY created_at DESC LIMIT $1", limit)
	}
	if err != nil {
		return nil, fmt.Errorf("alloydb: query: %w", err)
	}
	defer rows.Close()
	out := make([]AlertRecord, 0, 16)
	for rows.Next() {
		var a AlertRecord
		if err := rows.Scan(&a.AlertID, &a.TxnID, &a.CustomerID, &a.Severity, &a.Reason, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("alloydb: scan: %w", err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("alloydb: rows: %w", err)
	}
	return out, nil
}

// Select returns the appropriate Client for the runtime configuration.
//
//   - When cfg.AlloyDBDSN is set, a *RealClient is created against that
//     DSN and the env-var driven live path is active.
//   - Otherwise the in-process MockClient is returned, preserving the
//     existing behavior and the Simulated=true contract.
//
// The caller owns the returned *RealClient and is responsible for
// closing its connection pool during shutdown. For the MockClient no
// cleanup is required.
func Select(ctx context.Context, cfg *config.Config) Client {
	if cfg != nil && cfg.AlloyDBDSN != "" {
		c, err := NewRealClient(ctx, cfg.AlloyDBDSN)
		if err == nil {
			return c
		}
		// Fall back to mock on connect error so the service still boots;
		// the log line lets the operator see why live data is disabled.
		// We use fmt.Fprintln to stderr to avoid pulling in the project
		// logger here — alloydb has no log dependency.
		fmt.Fprintf(stderr, "alloydb: real client disabled (%v); using MockClient\n", err)
	}
	return MockClient{}
}

// stderr is overridable for tests.
var stderr = stdErr{}

type stdErr struct{}

func (stdErr) Write(p []byte) (int, error) { return fmt.Println(string(p)) }
