package main

import (
	"github.com/finguard/platform/pkg/domain"
	"github.com/finguard/platform/pkg/vertex"
)

// decisionEngine maps a prediction to a Decision + reasons.
type decisionEngine struct {
	ApproveMax float64 // < ApproveMax  → APPROVE
	BlockMin   float64 // >= BlockMin   → BLOCK
	// Between is REVIEW
}

func defaultDecisionEngine() *decisionEngine {
	return &decisionEngine{
		ApproveMax: 0.3,
		BlockMin:   0.7,
	}
}

// Decide returns a decision and the human-readable reasons.
func (d *decisionEngine) Decide(p vertex.Prediction) (domain.Decision, []string) {
	reasons := explain(p)
	switch {
	case p.FraudProbability >= d.BlockMin:
		return domain.DecisionBlock, reasons
	case p.FraudProbability >= d.ApproveMax:
		return domain.DecisionReview, reasons
	default:
		return domain.DecisionApprove, nil
	}
}

// explain returns top reasons based on the prediction inputs.
func explain(p vertex.Prediction) []string {
	if len(p.SHAPValues) == 0 {
		return []string{}
	}
	labels := []string{"amount_zscore", "geo_velocity", "device_risk", "merchant_risk"}
	// Pick the top 2 by magnitude
	type kv struct {
		idx   int
		value float64
	}
	var pairs []kv
	for i, v := range p.SHAPValues {
		pairs = append(pairs, kv{i, v})
	}
	// bubble-sort top 2 (small N)
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			if abs(pairs[j].value) > abs(pairs[i].value) {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}
	n := 2
	if n > len(pairs) {
		n = len(pairs)
	}
	reasons := make([]string, 0, n)
	for i := 0; i < n; i++ {
		idx := pairs[i].idx
		if idx < len(labels) {
			reasons = append(reasons, "unusual_"+labels[idx])
		}
	}
	if p.AnomalyScore > 0.6 {
		reasons = append(reasons, "anomaly_high")
	}
	return reasons
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
