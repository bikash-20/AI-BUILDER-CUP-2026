// Package riskstore is the read-side data layer for risk-service's
// /v1/customers/{id}/transactions endpoint.
//
// In production this would back onto BigQuery. The package exposes a
// single Client interface so the service can be wired against either:
//
//   - a deterministic in-process MockClient (for offline development and
//     unit tests), or
//   - a BigQueryClient that talks to the goccy BigQuery emulator over
//     HTTP for local end-to-end testing.
//
// Both implementations return a domain.CustomerTransactions envelope;
// the MockClient sets Simulated=true / Source="riskstore:MockClient",
// the BigQueryClient sets Simulated=false / Source="riskstore:BigQueryClient".
package riskstore

import (
	"context"

	"github.com/finguard/platform/pkg/domain"
)

// Client is the contract used by risk-service.
type Client interface {
	ListCustomerTransactions(ctx context.Context, customerID string, limit int) (domain.CustomerTransactions, error)
}
