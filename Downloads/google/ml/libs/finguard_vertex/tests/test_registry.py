"""Tests for finguard_vertex.registry."""

from __future__ import annotations

from finguard_vertex.registry import (
    ModelRegistry,
    deterministic_metrics,
    metrics_meet_threshold,
)


def test_register_and_resolve() -> None:
    reg = ModelRegistry()
    m = reg.register("fraud_xgb", "v1", metrics={"auc_roc": 0.85})
    reg.add_alias("fraud_xgb", "v1", "default")
    resolved = reg.resolve("fraud_xgb", "default")
    assert resolved is not None
    assert resolved.version == "v1"


def test_deterministic_metrics_shape() -> None:
    m = deterministic_metrics("run-123")
    assert "auc_roc" in m
    assert 0.0 <= m["auc_roc"] <= 1.0


def test_metrics_meet_threshold() -> None:
    m = {"auc_roc": 0.92, "f1": 0.7}
    assert metrics_meet_threshold(m, {"auc_roc": 0.9, "f1": 0.6}) is True
    assert metrics_meet_threshold(m, {"auc_roc": 0.95}) is False
