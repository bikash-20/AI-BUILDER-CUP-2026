"""Evaluation utilities: SHAP, fairness, and standard metrics."""

from finguard_evaluation.metrics import (
    binary_metrics,
    expected_calibration_error,
    population_stability_index,
)
from finguard_evaluation.shap_explain import ShapExplainer, compute_feature_importance

__all__ = [
    "binary_metrics",
    "compute_feature_importance",
    "expected_calibration_error",
    "population_stability_index",
    "ShapExplainer",
]
