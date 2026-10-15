"""Configuration for FinGuard Python services.

Loads env vars lazily via a cached `Settings` object. Centralizing this
makes it easy to inject mocks in tests via `finguard_common.config._reset_cache()`.
"""

from __future__ import annotations

import os
from dataclasses import dataclass, field
from functools import lru_cache
from typing import Any


@dataclass
class Settings:
    """Runtime configuration."""

    project_id: str
    region: str = "asia-south1"
    env: str = "dev"
    mock_mode: bool = True
    log_level: str = "INFO"

    # Vertex AI
    vertex_endpoint_id: str | None = None
    vertex_model_name: str = "fraud-ensemble"

    # BigQuery
    bigquery_dataset: str = "fin"
    bigquery_features_mv: str = "fin.customer_features_mv"

    # Pub/Sub
    pubsub_topic_scored: str = "txn-scored"
    pubsub_topic_alerts: str = "alerts"

    # Storage
    docai_bucket: str = "finguard-documents"
    dlq_bucket: str = "finguard-dlq"

    # Document AI
    docai_processor_id: str | None = None
    docai_location: str = "us"

    # Gemini
    gemini_model: str = "gemini-1.5-pro"
    vertex_search_datastore: str | None = None

    # Service URLs (downstream Go services in dev)
    fraud_service_url: str = "http://fraud-service:8080"
    risk_service_url: str = "http://risk-service:8080"
    chat_service_url: str = "http://chat-bff:8080"

    feature_flags: dict[str, bool] = field(default_factory=dict)


def _coerce_bool(value: str | None, default: bool = False) -> bool:
    if value is None:
        return default
    return value.strip().lower() in {"1", "true", "yes", "y", "on"}


def _parse_flags(raw: str | None) -> dict[str, bool]:
    out: dict[str, bool] = {}
    if not raw:
        return out
    for kv in raw.split(","):
        kv = kv.strip()
        if not kv:
            continue
        if "=" in kv:
            k, v = kv.split("=", 1)
            out[k.strip()] = _coerce_bool(v, True)
        else:
            out[kv] = True
    return out


def _load() -> Settings:
    project = os.getenv("GCP_PROJECT_ID", "finguard-dev")
    return Settings(
        project_id=project,
        region=os.getenv("GCP_REGION", "asia-south1"),
        env=os.getenv("ENV", "dev"),
        mock_mode=_coerce_bool(os.getenv("MOCK_MODE"), True),
        log_level=os.getenv("LOG_LEVEL", "INFO"),
        vertex_endpoint_id=os.getenv("VERTEX_ENDPOINT_ID"),
        vertex_model_name=os.getenv("VERTEX_MODEL_NAME", "fraud-ensemble"),
        bigquery_dataset=os.getenv("BIGQUERY_DATASET", "fin"),
        bigquery_features_mv=os.getenv("BIGQUERY_FEATURE_MV", "fin.customer_features_mv"),
        pubsub_topic_scored=os.getenv("PUBSUB_TOPIC_SCORED", "txn-scored"),
        pubsub_topic_alerts=os.getenv("PUBSUB_TOPIC_ALERTS", "alerts"),
        docai_bucket=os.getenv("DOCAI_BUCKET", "finguard-documents"),
        dlq_bucket=os.getenv("DLQ_BUCKET", "finguard-dlq"),
        docai_processor_id=os.getenv("DOCAI_PROCESSOR_ID"),
        docai_location=os.getenv("DOCAI_LOCATION", "us"),
        gemini_model=os.getenv("GEMINI_MODEL", "gemini-1.5-pro"),
        vertex_search_datastore=os.getenv("VERTEX_SEARCH_DATASTORE"),
        fraud_service_url=os.getenv("FRAUD_SERVICE_URL", "http://fraud-service:8080"),
        risk_service_url=os.getenv("RISK_SERVICE_URL", "http://risk-service:8080"),
        chat_service_url=os.getenv("CHAT_SERVICE_URL", "http://chat-bff:8080"),
        feature_flags=_parse_flags(os.getenv("FEATURE_FLAGS")),
    )


@lru_cache(maxsize=1)
def get_settings() -> Settings:
    """Return the cached Settings instance."""
    return _load()


def reset_cache() -> None:
    """Clear the cached settings (useful in tests)."""
    get_settings.cache_clear()


def is_mock_mode() -> bool:
    return get_settings().mock_mode


def project_id() -> str:
    return get_settings().project_id


def region() -> str:
    return get_settings().region
