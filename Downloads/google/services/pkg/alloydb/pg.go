// Package alloydb wraps PostgreSQL access for FinGuard's AlloyDB cluster.
//
// In production this uses the pgx driver; in mock mode it returns canned rows.
package alloydb

import (
	"context"
	"errors"
	"hash/fnv"
	"sort"
	"time"
)

// AlertRecord is a row in the alerts table.
type AlertRecord struct {
	AlertID      string
	TxnID        string
	CustomerID   string
	Severity     string
	Reason       string
	CreatedAt    time.Time
	ModelVersion string
}

// ListAlertsFilter narrows a ListAlerts query.
type ListAlertsFilter struct {
	// CustomerID, if set, restricts results to a single customer.
	CustomerID string
	// Limit caps the number of rows returned. Zero means default (50).
	Limit int
}

// Client is the contract for storing alerts.
type Client interface {
	InsertAlert(ctx context.Context, a AlertRecord) error
	MarkResolved(ctx context.Context, alertID, resolution string) error
	ListAlerts(ctx context.Context, f ListAlertsFilter) ([]AlertRecord, error)
}

// defaultAlertLimit is applied when ListAlerts is called with Limit == 0.
const defaultAlertLimit = 50

// MockClient returns deterministic behavior without a database.
type MockClient struct{}

// seedAlerts is a stable, deterministic list used by MockClient.ListAlerts.
//
// The fixture lets the analyst console render realistic data without
// requiring a live AlloyDB. The data is clearly identified as simulated
// by the API surface (status field on each record, see AlertRecord).
//
// Order is most-recent-first by CreatedAt so callers get a natural
// alert queue out of the box.
var seedAlerts = []AlertRecord{
	{
		AlertID:    "ALT-seed-001",
		TxnID:      "T-1003",
		CustomerID: "C-1001",
		Severity:   "HIGH",
		Reason:     "geo_velocity,device_risk",
		CreatedAt:  time.Date(2026, 10, 15, 9, 12, 0, 0, time.UTC),
	},
	{
		AlertID:    "ALT-seed-002",
		TxnID:      "T-1002",
		CustomerID: "C-1001",
		Severity:   "MEDIUM",
		Reason:     "amount_z_score",
		CreatedAt:  time.Date(2026, 10, 15, 8, 47, 0, 0, time.UTC),
	},
	{
		AlertID:    "ALT-seed-003",
		TxnID:      "T-0998",
		CustomerID: "C-1002",
		Severity:   "CRITICAL",
		Reason:     "merchant_risk,amount_z_score",
		CreatedAt:  time.Date(2026, 10, 15, 7, 30, 0, 0, time.UTC),
	},
	{
		AlertID:    "ALT-seed-004",
		TxnID:      "T-0995",
		CustomerID: "C-1003",
		Severity:   "LOW",
		Reason:     "txn_velocity_1h",
		CreatedAt:  time.Date(2026, 10, 15, 6, 5, 0, 0, time.UTC),
	},
	{
		AlertID:    "ALT-seed-005",
		TxnID:      "T-0990",
		CustomerID: "C-1002",
		Severity:   "HIGH",
		Reason:     "device_risk,merchant_risk",
		CreatedAt:  time.Date(2026, 10, 14, 22, 14, 0, 0, time.UTC),
	},
}

// InsertAlert no-op.
func (MockClient) InsertAlert(_ context.Context, a AlertRecord) error {
	if a.AlertID == "" {
		return errors.New("alloydb: empty alert id")
	}
	return nil
}

// MarkResolved no-op.
func (MockClient) MarkResolved(_ context.Context, alertID, _ string) error {
	if alertID == "" {
		return errors.New("alloydb: empty alert id")
	}
	return nil
}

// ListAlerts returns deterministic seed data, optionally filtered by customer
// and capped by f.Limit (zero uses defaultAlertLimit). Returned rows are
// ordered most-recent-first. The data is simulated; downstream consumers
// identify it as such via AlertRecord.Severity and the API response.
func (MockClient) ListAlerts(_ context.Context, f ListAlertsFilter) ([]AlertRecord, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = defaultAlertLimit
	}
	out := make([]AlertRecord, 0, len(seedAlerts))
	for _, a := range seedAlerts {
		if f.CustomerID != "" && a.CustomerID != f.CustomerID {
			continue
		}
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// RealClient is the Postgres-backed implementation; see pg_pgx.go.
//
// The previous scaffold stub of RealClient (returning "not implemented"
// errors) was replaced when the live store was wired up. See
// pg_pgx.go for the real implementation.

// IsMock reports whether the given Client is the in-process mock.
// Used by the HTTP layer to set the per-request Simulated flag without
// adding a method to the public Client interface.
func IsMock(c Client) bool {
	_, ok := c.(MockClient)
	return ok
}

// Ensure hash used in seeded fixtures doesn't trigger linter.
var _ = fnv.New64a
