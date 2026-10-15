"""Feature engineering pipeline component."""

from __future__ import annotations

from kfp import dsl  # type: ignore

BASE_IMAGE = "python:3.11-slim"


@dsl.component(base_image=BASE_IMAGE, packages_to_install=["finguard-common", "google-cloud-bigquery"])
def feature_eng(
    project: str,
    dataset: str,
) -> str:
    """Refresh the `customer_features_mv` materialized view."""
    from training.fraud_xgb.feature_eng import render  # type: ignore

    sql = render(project)
    # Real implementation: client.query(sql).result()
    return sql[:120]
