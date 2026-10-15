package main

import (
	"context"

	"github.com/finguard/platform/pkg/bigquery"
)

// enricherAdapter adapts the BQ feature client to a typed shape used in tests.
type enricherAdapter struct {
	c bigquery.FeatureClient
}

// GetCustomerFeatures returns features for fraud-service.
func (e enricherAdapter) GetCustomerFeatures(ctx context.Context, customerID string) (*bigquery.FeatureRow, error) {
	return e.c.GetCustomerFeatures(ctx, customerID)
}
