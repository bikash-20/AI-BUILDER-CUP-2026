"""Tests for finguard_vertex.predictor."""

from __future__ import annotations

from finguard_vertex.predictor import (
    EnsemblePredictor,
    RuleBasedPredictor,
    stable_score,
)


def test_rule_based_predictor_deterministic() -> None:
    p = RuleBasedPredictor()
    a = p.predict({"amount_zscore": 0.4, "geo_velocity": 0.2})
    b = p.predict({"amount_zscore": 0.4, "geo_velocity": 0.2})
    assert a.fraud_probability == b.fraud_probability


def test_rule_based_predictor_range() -> None:
    p = RuleBasedPredictor()
    for _ in range(50):
        r = p.predict({})
        assert 0.0 <= r.fraud_probability <= 1.0


def test_ensemble_average() -> None:
    a = RuleBasedPredictor()
    e = EnsemblePredictor(a, a, weights=[0.6, 0.4])
    r = e.predict({"amount_zscore": 0.1})
    assert 0.0 <= r.fraud_probability <= 1.0


def test_stable_score_range() -> None:
    for s in ["a", "b", "c", "long string with spaces and !@#"]:
        v = stable_score(s)
        assert 0.0 <= v <= 1.0
