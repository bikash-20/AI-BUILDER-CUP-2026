package main

import (
	"testing"

	"github.com/finguard/platform/pkg/domain"
	"github.com/finguard/platform/pkg/vertex"
)

func TestDecideApprove(t *testing.T) {
	d := defaultDecisionEngine()
	got, reasons := d.Decide(vertex.Prediction{FraudProbability: 0.1})
	if got != domain.DecisionApprove {
		t.Errorf("got %v, want APPROVE", got)
	}
	if reasons != nil {
		t.Errorf("expected nil reasons, got %v", reasons)
	}
}

func TestDecideReview(t *testing.T) {
	d := defaultDecisionEngine()
	got, _ := d.Decide(vertex.Prediction{FraudProbability: 0.5})
	if got != domain.DecisionReview {
		t.Errorf("got %v, want REVIEW", got)
	}
}

func TestDecideBlock(t *testing.T) {
	d := defaultDecisionEngine()
	got, _ := d.Decide(vertex.Prediction{FraudProbability: 0.9})
	if got != domain.DecisionBlock {
		t.Errorf("got %v, want BLOCK", got)
	}
}

func TestExplain(t *testing.T) {
	d := defaultDecisionEngine()
	_, reasons := d.Decide(vertex.Prediction{
		FraudProbability: 0.9,
		AnomalyScore:     0.7,
		SHAPValues:       []float64{0.1, 0.5, 0.3, 0.05},
	})
	if len(reasons) == 0 {
		t.Fatal("expected reasons")
	}
	found := false
	for _, r := range reasons {
		if r == "unusual_geo_velocity" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected unusual_geo_velocity in reasons, got %v", reasons)
	}
}
