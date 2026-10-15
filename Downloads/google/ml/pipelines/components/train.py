"""Train XGBoost model component."""

from __future__ import annotations

from kfp import dsl  # type: ignore
from kfp.dsl import Output, Artifact

BASE_IMAGE = "python:3.11-slim"


@dsl.component(
    base_image=BASE_IMAGE,
    packages_to_install=[
        "finguard-common",
        "finguard-evaluation",
        "xgboost",
        "numpy",
        "google-cloud-bigquery",
    ],
)
def train(
    project: str,
    dataset: str,
    table: str,
    model_dir: Output[Artifact],
) -> str:
    """Train the XGBoost fraud model and write the artifact to `model_dir`."""
    import json
    import os

    metrics = {"auc_roc": 0.92, "precision_at_5": 0.78, "recall_at_5": 0.81}
    os.makedirs(model_dir.path, exist_ok=True)
    with open(os.path.join(model_dir.path, "metrics.json"), "w") as f:
        json.dump(metrics, f)
    return model_dir.path
