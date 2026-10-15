"""SHAP explainability helpers.

Wraps shap.TreeExplainer / Generic explainers behind a uniform interface.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any

from finguard_common.logging import get_logger

log = get_logger("finguard.evaluation.shap")


@dataclass
class ShapExplainer:
    """Compute SHAP values for a single prediction.

    If `shap` is unavailable, falls back to a permutation-based estimator.
    """

    model: Any
    feature_names: list[str] = field(default_factory=list)
    background: Any | None = None

    def explain(self, instance: dict[str, float]) -> dict[str, float]:
        try:
            import shap  # type: ignore

            explainer = shap.Explainer(self.model)
            values = explainer([list(instance.values())])
            arr = values.values[0] if hasattr(values, "values") else values[0]
            names = self.feature_names or list(instance.keys())
            return {names[i]: float(arr[i]) for i in range(len(names))}
        except Exception as e:
            log.warning("shap_unavailable", error=str(e))
            return _permutation_importance(self.model, instance)


def compute_feature_importance(model: Any, X: list[dict[str, float]], y: list[int]) -> dict[str, float]:
    """Rank features by simple permutation importance (mean abs decrease in score)."""
    base = _predict(model, X)
    out: dict[str, float] = {}
    feature_names = list(X[0].keys()) if X else []
    for feat in feature_names:
        perturbed = [dict(row) for row in X]
        for row in perturbed:
            row[feat] = 0.0
        new = _predict(model, perturbed)
        out[feat] = sum(abs(b - n) for b, n in zip(base, new)) / max(len(base), 1)
    return out


def _predict(model: Any, X: list[dict[str, float]]) -> list[float]:
    if hasattr(model, "predict_proba"):
        proba = model.predict_proba(X)
        return [float(p[1]) for p in proba]
    if hasattr(model, "predict"):
        return [float(v) for v in model.predict(X)]
    return [0.5 for _ in X]


def _permutation_importance(model: Any, instance: dict[str, float]) -> dict[str, float]:
    base = _predict(model, [instance])[0]
    out: dict[str, float] = {}
    for k, v in instance.items():
        new = dict(instance)
        new[k] = 0.0
        s = _predict(model, [new])[0]
        out[k] = base - s
    return out
