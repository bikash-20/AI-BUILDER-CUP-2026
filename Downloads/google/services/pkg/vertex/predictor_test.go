package vertex

import (
	"context"
	"testing"
)

func TestMockPredictorStableForKey(t *testing.T) {
	p := &MockPredictor{ModelVersion: "v1"}
	f := Features{AmountZScore: 0.1, GeoVelocity: 0.2, MerchantRisk: 0.3}
	a, _ := p.Predict(context.Background(), f, "txn-1")
	b, _ := p.Predict(context.Background(), f, "txn-1")
	if a.FraudProbability != b.FraudProbability {
		t.Errorf("not stable: %v vs %v", a.FraudProbability, b.FraudProbability)
	}
}

func TestMockPredictorProbabilityInRange(t *testing.T) {
	p := &MockPredictor{}
	for i := 0; i < 50; i++ {
		key := string(rune('a' + i))
		out, err := p.Predict(context.Background(), Features{}, key)
		if err != nil {
			t.Fatalf("Predict error: %v", err)
		}
		if out.FraudProbability < 0 || out.FraudProbability > 1 {
			t.Errorf("out of range: %v", out.FraudProbability)
		}
	}
}

func TestMockPredictorFailureRate(t *testing.T) {
	p := &MockPredictor{FailureRate: 1.0}
	_, err := p.Predict(context.Background(), Features{}, "x")
	if err == nil {
		t.Error("expected error when FailureRate=1")
	}
}
