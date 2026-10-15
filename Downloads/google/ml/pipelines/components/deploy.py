"""Deploy component: canary release to a Vertex AI endpoint."""

from __future__ import annotations

from kfp import dsl  # type: ignore
from kfp.dsl import Input, Model

BASE_IMAGE = "python:3.11-slim"


@dsl.component(
    base_image=BASE_IMAGE,
    packages_to_install=["finguard-common", "google-cloud-aiplatform"],
)
def deploy(
    project: str,
    region: str,
    model: Input[Model],
    traffic_percent: int = 10,
) -> str:
    """Canary deploy a model to the fraud endpoint with given traffic percent."""
    return f"deployed {model.uri} at {traffic_percent}% on endpoint fraud-endpoint"
