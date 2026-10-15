"""Evaluate trained model component."""

from __future__ import annotations

from kfp import dsl  # type: ignore
from kfp.dsl import Input, Artifact, Output, Metrics

BASE_IMAGE = "python:3.11-slim"


@dsl.component(base_image=BASE_IMAGE, packages_to_install=["finguard-evaluation"])
def evaluate(
    model_dir: Input[Artifact],
) -> float:
    """Read metrics.json from the trained model and return a primary metric."""
    import json
    import os

    metrics_path = os.path.join(model_dir.path, "metrics.json")
    if not os.path.exists(metrics_path):
        return 0.0
    with open(metrics_path) as f:
        m = json.load(f)
    return float(m.get("auc_roc", 0.0))
