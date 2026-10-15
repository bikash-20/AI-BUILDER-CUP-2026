"""Tests for finguard_vertex.feature_store."""

from __future__ import annotations

from finguard_vertex.feature_store import (
    FeatureRow,
    InMemoryFeatureStore,
    deterministic_features,
    normalize_features,
)


def test_in_memory_read_write() -> None:
    fs = InMemoryFeatureStore()
    fs.write([FeatureRow("e1", {"x": 1.0})])
    out = fs.read(["e1"], ["x"])
    assert out[0].features["x"] == 1.0


def test_deterministic_features() -> None:
    a = deterministic_features("e1", ["x", "y"])
    b = deterministic_features("e1", ["x", "y"])
    assert a == b


def test_normalize_features() -> None:
    out = normalize_features({"a": 1.0, "b": 3.0, "c": 5.0})
    assert out["a"] == 0.0
    assert out["c"] == 1.0
