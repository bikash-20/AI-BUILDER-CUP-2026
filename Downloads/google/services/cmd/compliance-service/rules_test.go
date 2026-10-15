package main

import "testing"

func TestRulesEvaluate(t *testing.T) {
	r := NewRules()
	if !r.Evaluate(10_000, 0) {
		t.Error("expected trigger on amount >= 10k")
	}
	if !r.Evaluate(100, 3) {
		t.Error("expected trigger on suspicious countries")
	}
	if r.Evaluate(100, 1) {
		t.Error("expected no trigger")
	}
}
