"""Vertex AI Model Registry helpers.

In production this would manage models through `aiplatform.models.Model` and
`aiplatform.models.ModelRegistry`. In mock mode it tracks versions in-process.
"""

from __future__ import annotations

import hashlib
import time
from dataclasses import dataclass, field
from typing import Any

from finguard_common.logging import get_logger

log = get_logger("finguard.registry")


@dataclass
class ModelVersion:
    model_id: str
    version: str
    metrics: dict[str, float] = field(default_factory=dict)
    created_at: float = field(default_factory=time.time)
    aliases: list[str] = field(default_factory=list)


@dataclass
class ModelRegistry:
    """In-process registry, swap with aiplatform.models.ModelRegistry in prod."""

    _models: dict[str, list[ModelVersion]] = field(default_factory=dict)

    def register(self, model_id: str, version: str, metrics: dict[str, float]) -> ModelVersion:
        """Add a new version of a model."""
        v = ModelVersion(model_id=model_id, version=version, metrics=metrics)
        self._models.setdefault(model_id, []).append(v)
        log.info("model_registered", model_id=model_id, version=version)
        return v

    def add_alias(self, model_id: str, version: str, alias: str) -> None:
        for v in self._models.get(model_id, []):
            if v.version == version:
                if alias not in v.aliases:
                    v.aliases.append(alias)
                return
        raise KeyError(f"version {version} not found for {model_id}")

    def resolve(self, model_id: str, alias: str = "default") -> ModelVersion | None:
        for v in self._models.get(model_id, []):
            if alias in v.aliases:
                return v
        return None

    def versions(self, model_id: str) -> list[str]:
        return [v.version for v in self._models.get(model_id, [])]


def deterministic_metrics(seed: str) -> dict[str, float]:
    """Return deterministic evaluation metrics for a given run id."""
    h = hashlib.sha256(seed.encode()).digest()
    return {
        "auc_roc": (h[0] / 255.0) * 0.4 + 0.6,  # 0.6..1.0
        "precision_at_5": (h[1] / 255.0) * 0.5 + 0.5,
        "recall_at_5": (h[2] / 255.0) * 0.5 + 0.5,
        "f1": (h[3] / 255.0) * 0.3 + 0.6,
    }


def metrics_meet_threshold(metrics: dict[str, float], threshold: dict[str, float]) -> bool:
    """Check whether metrics meet a minimum threshold set (e.g., for promotion)."""
    for k, v in threshold.items():
        if metrics.get(k, 0.0) < v:
            return False
    return True
