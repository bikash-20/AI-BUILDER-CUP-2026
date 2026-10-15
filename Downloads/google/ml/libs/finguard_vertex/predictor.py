"""Vertex AI online prediction clients.

In production, `VertexOnlinePredictor` calls `aiplatform.Endpoint.predict(...)`.
In mock mode, `RuleBasedPredictor` returns a deterministic probability for tests.
"""

from __future__ import annotations

import hashlib
import math
import time
from dataclasses import dataclass
from typing import Any, Protocol

from finguard_common.config import is_mock_mode
from finguard_common.logging import get_logger

log = get_logger("finguard.vertex")


@dataclass
class Prediction:
    fraud_probability: float
    anomaly_score: float
    shap_values: list[float]
    model_version: str

    def to_dict(self) -> dict[str, Any]:
        return {
            "fraud_probability": self.fraud_probability,
            "anomaly_score": self.anomaly_score,
            "shap_values": self.shap_values,
            "model_version": self.model_version,
        }


class Predictor(Protocol):
    def predict(self, instance: dict[str, Any]) -> Prediction: ...


class RuleBasedPredictor:
    """Deterministic fallback used in tests / mock mode."""

    model_version = "rule-v0"

    def predict(self, instance: dict[str, Any]) -> Prediction:
        z = float(instance.get("amount_zscore", 0.0))
        gv = float(instance.get("geo_velocity", 0.0))
        dr = float(instance.get("device_risk", 0.0))
        mr = float(instance.get("merchant_risk", 0.0))
        s = 0.5 * z + 0.3 * gv + 0.15 * dr + 0.05 * mr
        s = max(0.0, min(1.0, s + 0.5))  # sigmoid-ish bias
        return Prediction(
            fraud_probability=s,
            anomaly_score=gv,
            shap_values=[0.5 * z, 0.3 * gv, 0.15 * dr, 0.05 * mr],
            model_version=self.model_version,
        )


class VertexOnlinePredictor:
    """Calls a deployed Vertex AI endpoint.

    Usage::

        p = VertexOnlinePredictor(endpoint_id="123", project="x", region="asia-south1")
        p.predict({"amount_zscore": 0.1})
    """

    def __init__(self, endpoint_id: str, project: str, region: str = "asia-south1") -> None:
        self.endpoint_id = endpoint_id
        self.project = project
        self.region = region
        self._endpoint = None
        if not is_mock_mode():
            try:
                from google.cloud import aiplatform  # type: ignore

                aiplatform.init(project=project, location=region)
                self._endpoint = aiplatform.Endpoint(endpoint_id)
            except Exception as e:  # pragma: no cover
                log.warning("vertex_init_failed", error=str(e))
                self._endpoint = None

    def predict(self, instance: dict[str, Any]) -> Prediction:
        if self._endpoint is None:
            return RuleBasedPredictor().predict(instance)
        start = time.perf_counter()
        try:
            response = self._endpoint.predict(instances=[instance])
            preds = response.predictions[0]
            return Prediction(
                fraud_probability=float(preds.get("fraud_probability", 0.5)),
                anomaly_score=float(preds.get("anomaly_score", 0.0)),
                shap_values=list(preds.get("shap_values", [])),
                model_version=str(preds.get("model_version", "v0")),
            )
        except Exception as e:
            log.warning("vertex_predict_failed", error=str(e), latency=time.perf_counter() - start)
            return RuleBasedPredictor().predict(instance)


class EnsemblePredictor:
    """Average of two predictors; useful when there are two model versions live."""

    def __init__(self, *predictors: Predictor, weights: list[float] | None = None) -> None:
        if not predictors:
            raise ValueError("EnsemblePredictor requires at least one predictor")
        self.predictors = predictors
        if weights is None:
            self.weights = [1.0 / len(predictors)] * len(predictors)
        else:
            if len(weights) != len(predictors):
                raise ValueError("weights length must match predictors")
            s = sum(weights)
            self.weights = [w / s for w in weights]

    def predict(self, instance: dict[str, Any]) -> Prediction:
        preds = [p.predict(instance) for p in self.predictors]
        avg_prob = sum(p.fraud_probability * w for p, w in zip(preds, self.weights))
        avg_anom = sum(p.anomaly_score * w for p, w in zip(preds, self.weights))
        return Prediction(
            fraud_probability=avg_prob,
            anomaly_score=avg_anom,
            shap_values=preds[0].shap_values,
            model_version=preds[0].model_version,
        )


def stable_score(seed: str) -> float:
    """Deterministic 0..1 score from a string seed (used by tests)."""
    h = int(hashlib.sha1(seed.encode()).hexdigest()[:8], 16)
    return (h % 10_000) / 10_000.0


def sigmoid(x: float) -> float:
    return 1.0 / (1.0 + math.exp(-x))
