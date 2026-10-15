"""Smoke tests for pipeline components."""

from __future__ import annotations

import importlib

COMPONENTS = [
    "pipelines.components.data_validation",
    "pipelines.components.deploy",
    "pipelines.components.evaluate",
    "pipelines.components.feature_eng",
    "pipelines.components.register",
    "pipelines.components.train",
]


def test_components_importable() -> None:
    for name in COMPONENTS:
        mod = importlib.import_module(name)
        assert mod is not None
