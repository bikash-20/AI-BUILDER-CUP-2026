"""Tests for finguard_evaluation.metrics."""

from __future__ import annotations

import pytest

from finguard_evaluation.metrics import (
    binary_metrics,
    expected_calibration_error,
    population_stability_index,
)


def test_binary_metrics_perfect() -> None:
    m = binary_metrics([1, 1, 0, 0], [1.0, 1.0, 0.0, 0.0])
    assert m["precision"] == 1.0
    assert m["recall"] == 1.0
    assert m["f1"] == 1.0
    assert m["auc_roc"] == 1.0


def test_binary_metrics_handles_empty() -> None:
    m = binary_metrics([], [])
    assert m == {"precision": 0.0, "recall": 0.0, "f1": 0.0, "auc_roc": 0.0}


def test_binary_metrics_length_mismatch() -> None:
    with pytest.raises(ValueError):
        binary_metrics([1, 0], [0.5])


def test_ece_zero_when_well_calibrated() -> None:
    # All probs around 0.9 with truth 1: tight distribution, ECE near 0
    y = [1, 1, 1, 1, 0, 0]
    s = [0.9, 0.9, 0.9, 0.9, 0.1, 0.1]
    ece = expected_calibration_error(y, s)
    assert ece < 0.1


def test_psi_zero_for_same_distribution() -> None:
    s = [0.1, 0.2, 0.3, 0.4, 0.5, 0.6]
    assert population_stability_index(s, s) < 0.01


def test_psi_high_for_shifted_distribution() -> None:
    expected = [0.1, 0.2, 0.3]
    actual = [0.7, 0.8, 0.9]
    assert population_stability_index(expected, actual) > 0.5
