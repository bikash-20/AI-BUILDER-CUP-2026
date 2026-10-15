// Package bigquery wraps the Cloud BigQuery client for feature lookups.
package bigquery

import (
	"context"
	"errors"
	"hash/fnv"
	"math"
)

// FeatureRow is the per-customer feature vector used by fraud-service.
type FeatureRow struct {
	CustomerID      string
	TxnVelocity1h   float64
	TxnVelocity24h  float64
	GeoVelocity     float64
	DeviceRisk      float64
	MerchantRisk    float64
	AmountZScore    float64
}

// FeatureClient is the contract for feature lookups.
type FeatureClient interface {
	GetCustomerFeatures(ctx context.Context, customerID string) (*FeatureRow, error)
}

// MockClient returns deterministic stub features.
type MockClient struct{}

// GetCustomerFeatures hashes the customer id to produce stable pseudo-features.
func (MockClient) GetCustomerFeatures(_ context.Context, customerID string) (*FeatureRow, error) {
	if customerID == "" {
		return nil, errors.New("bigquery: empty customer id")
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(customerID))
	v := float64(h.Sum64()%1000) / 1000.0
	return &FeatureRow{
		CustomerID:     customerID,
		TxnVelocity1h:  math.Mod(v*7, 5),
		TxnVelocity24h: math.Mod(v*11, 20),
		GeoVelocity:    math.Mod(v*13, 1),
		DeviceRisk:     math.Mod(v*17, 1),
		MerchantRisk:   math.Mod(v*19, 1),
		AmountZScore:   v - 0.5,
	}, nil
}

// Client is the real implementation; stubbed here.
// In production it would call BigQuery's feature MV and return real values.
type Client struct {
	Project   string
	Dataset   string
	FeatureMV string
}

// GetCustomerFeatures is unimplemented in scaffold; use MockClient.
func (c *Client) GetCustomerFeatures(ctx context.Context, customerID string) (*FeatureRow, error) {
	return nil, errors.New("bigquery: Client.GetCustomerFeatures not implemented (use MockClient)")
}
