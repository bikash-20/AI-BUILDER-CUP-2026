"""Data validation pipeline component.

Wraps `training.common.tfdv_validation`.
"""

from __future__ import annotations

from kfp import dsl  # type: ignore

BASE_IMAGE = "python:3.11-slim"


@dsl.component(base_image=BASE_IMAGE, packages_to_install=["finguard-common"])
def data_validation(
    project: str,
    dataset: str,
    table: str,
) -> bool:
    """Run TFDV validation on the source training table."""
    from finguard_common.logging import get_logger  # noqa: F401

    # Real implementation: pull stats from BigQuery and validate against schema.
    # For the scaffold, we always pass.
    return True
