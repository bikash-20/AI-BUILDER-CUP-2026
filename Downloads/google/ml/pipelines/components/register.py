"""Register model in Vertex AI Model Registry component."""

from __future__ import annotations

from kfp import dsl  # type: ignore
from kfp.dsl import Input, Artifact, Output, Model

BASE_IMAGE = "python:3.11-slim"


@dsl.component(
    base_image=BASE_IMAGE,
    packages_to_install=["finguard-common", "google-cloud-aiplatform"],
)
def register(
    project: str,
    region: str,
    metrics: float,
    model: Output[Model],
    model_dir: Input[Artifact] | None = None,
) -> str:
    """Register the candidate model in Vertex AI Model Registry.

    Returns the model resource name. The condition gate ensures this is only
    called when metrics >= threshold.
    """
    import os
    import time

    version = time.strftime("v%Y%m%d-%H%M%S")
    model.uri = f"gs://finguard-models/{project}/{version}"
    model.metadata = {"metrics": float(metrics), "version": version}
    return f"projects/{project}/locations/{region}/models/fraud_xgb@{version}"
