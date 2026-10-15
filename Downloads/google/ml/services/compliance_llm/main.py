"""Entrypoint for compliance_llm service."""

from __future__ import annotations

from finguard_common.logging import configure_logging

configure_logging()
from app import app  # noqa: E402

__all__ = ["app"]
