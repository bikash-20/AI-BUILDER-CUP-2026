package main

import (
	"context"
	"errors"

	"github.com/finguard/platform/pkg/vertex"
)

// ruleBasedScorer is a deterministic fallback used when Vertex AI is unavailable
// or returns an error.
type ruleBasedScorer struct{}

func (ruleBasedScorer) Predict(_ context.Context, f vertex.Features, _ string) (*vertex.Prediction, error) {
	if f.AmountZScore == 0 && f.GeoVelocity == 0 && f.DeviceRisk == 0 {
		return nil, errors.New("scorer: no features")
	}
	// Simple weighted sum into a probability.
	score := 0.5*f.AmountZScore + 0.3*f.GeoVelocity + 0.2*f.DeviceRisk + 0.1*f.MerchantRisk
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	return &vertex.Prediction{
		FraudProbability: score,
		AnomalyScore:     f.GeoVelocity,
		ModelVersion:     "rule-v0",
		SHAPValues: []float64{
			0.5 * f.AmountZScore,
			0.3 * f.GeoVelocity,
			0.2 * f.DeviceRisk,
			0.1 * f.MerchantRisk,
		},
	}, nil
}
