"""Tests for finguard_evaluation.shap_explain (using permutation fallback)."""

from __future__ import annotations

from finguard_evaluation.shap_explain import compute_feature_importance


class DummyModel:
    def predict(self, X):
        # Always returns the "a" feature value
        return [float(row.get("a", 0.0)) for row in X]


def test_compute_feature_importance() -> None:
    X = [{"a": 1.0, "b": 2.0}, {"a": 0.0, "b": 1.0}, {"a": 1.0, "b": 0.0}]
    y = [1, 0, 1]
    out = compute_feature_importance(DummyModel(), X, y)
    assert "a" in out
    assert "b" in out
    # "a" matters, "b" doesn't, so importance of a > b
    assert out["a"] > out["b"]
