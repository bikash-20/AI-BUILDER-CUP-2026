"""Entrypoint for the gemini_service Cloud Run service.

Run with:
    MOCK_MODE=true uvicorn app:app --host 0.0.0.0 --port 8080
"""

from __future__ import annotations

from finguard_common.logging import configure_logging

configure_logging()
from app import app  # noqa: E402

__all__ = ["app"]
