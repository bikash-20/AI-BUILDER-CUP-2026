"""Autoencoder evaluation utilities."""

from __future__ import annotations

import json

from finguard_evaluation.metrics import binary_metrics


def evaluate(mse_scores: list[float], labels: list[int], threshold: float | None = None) -> dict[str, float]:
    if threshold is None:
        threshold = sum(mse_scores) / len(mse_scores) if mse_scores else 0.0
    y_score = [1.0 if s > threshold else 0.0 for s in mse_scores]
    m = binary_metrics(labels, y_score)
    m["threshold"] = threshold
    return m


if __name__ == "__main__":
    import sys

    if len(sys.argv) < 2:
        print("usage: evaluate.py <input.json>")
        sys.exit(1)
    with open(sys.argv[1]) as f:
        data = json.load(f)
    print(json.dumps(evaluate(data["mse_scores"], data["labels"]), indent=2))
