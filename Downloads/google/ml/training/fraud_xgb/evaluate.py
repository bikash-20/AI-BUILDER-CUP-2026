"""Offline evaluation utilities for fraud XGBoost model."""

from __future__ import annotations

import json
import sys

from finguard_common.logging import configure_logging, get_logger
from finguard_evaluation.metrics import binary_metrics, expected_calibration_error, population_stability_index

configure_logging()
log = get_logger("finguard.training.fraud_xgb.evaluate")


def evaluate(y_true: list[int], y_score: list[float], reference: list[float] | None = None) -> dict[str, float]:
    out = binary_metrics(y_true, y_score)
    out["ece"] = expected_calibration_error(y_true, y_score)
    if reference is not None:
        out["psi"] = population_stability_index(reference, y_score)
    return out


def main() -> int:
    # CLI: pass a JSON file with {"y_true": [...], "y_score": [...], "reference": [...]}
    if len(sys.argv) < 2:
        log.error("usage: evaluate.py <input.json>")
        return 1
    with open(sys.argv[1]) as f:
        data = json.load(f)
    metrics = evaluate(data["y_true"], data["y_score"], data.get("reference"))
    print(json.dumps(metrics, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
