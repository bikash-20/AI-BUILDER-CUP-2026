// Package vertex wraps the Vertex AI online prediction client.
//
// In mock mode, the Predictor returns a deterministic pseudo-random score.
package vertex

import (
	"context"
	"errors"
	"hash/fnv"
	"math"
	"time"
)

// Features is the input feature vector for the fraud ensemble.
// All fields are optional; the consumer may omit any subset.
type Features struct {
	TxnVelocity1h  float64 `json:"txn_velocity_1h"`
	TxnVelocity24h float64 `json:"txn_velocity_24h"`
	GeoVelocity    float64 `json:"geo_velocity"`
	DeviceRisk     float64 `json:"device_risk"`
	MerchantRisk   float64 `json:"merchant_risk"`
	AmountZScore   float64 `json:"amount_zscore"`
	TimeOfDay      float64 `json:"time_of_day"`
	IsNewDevice    bool    `json:"is_new_device"`
	IsForeignGeo   bool    `json:"is_foreign_geo"`
}

// Prediction is the score returned by the fraud model.
type Prediction struct {
	FraudProbability float64   `json:"fraud_probability"`
	AnomalyScore     float64   `json:"anomaly_score"`
	SHAPValues       []float64 `json:"shap_values"`
	ModelVersion     string    `json:"model_version"`
}

// Predictor is the contract for fraud prediction implementations.
type Predictor interface {
	Predict(ctx context.Context, f Features, instanceKey string) (*Prediction, error)
}

// MockPredictor is a deterministic fallback used in local development.
type MockPredictor struct {
	ModelVersion string
	FailureRate  float64 // 0..1
}

// Predict returns a hash-based pseudo-fraud probability.
func (m *MockPredictor) Predict(_ context.Context, f Features, instanceKey string) (*Prediction, error) {
	if m.FailureRate > 0 && stableRand(instanceKey) < m.FailureRate {
		return nil, errors.New("mock predictor failure")
	}
	h := fnv.New64a()
	h.Write([]byte(instanceKey))
	score := float64(h.Sum64()%1000) / 1000.0
	score = 1/(1+math.Exp(-(score*4-2+0.5*f.AmountZScore+0.3*f.GeoVelocity))) // sigmoid

	return &Prediction{
		FraudProbability: clamp01(score),
		AnomalyScore:     clamp01(f.GeoVelocity*0.4 + f.AmountZScore*0.2),
		SHAPValues: []float64{
			0.3 * f.AmountZScore,
			0.25 * f.GeoVelocity,
			0.2 * f.DeviceRisk,
			0.15 * f.MerchantRisk,
		},
		ModelVersion: orDefault(m.ModelVersion, "mock-v1"),
	}, nil
}

// ClientPredictor is the production implementation that calls Vertex AI.
// Implementation is intentionally stubbed to keep the scaffold hermetic.
type ClientPredictor struct {
	EndpointID  string
	ProjectID    string
	Region       string
	ModelVersion string
}

// Predict calls Vertex AI; unimplemented in scaffold — use MockPredictor.
func (c *ClientPredictor) Predict(ctx context.Context, _ Features, _ string) (*Prediction, error) {
	return nil, errors.New("vertex: ClientPredictor.Predict not implemented (use MockPredictor)")
}

// Helpers ----

func clamp01(v float64) float64 {
	return math.Min(1, math.Max(0, v))
}

func orDefault(v, d string) string {
	if v == "" {
		return d
	}
	return v
}

// stableRand returns a deterministic 0..1 value for a string.
func stableRand(s string) float64 {
	if s == "" {
		return 0
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return float64(h.Sum64()%1000) / 1000.0
}

// ensure time package is used in the scaffold even when ClientPredictor is empty
var _ = time.Now
