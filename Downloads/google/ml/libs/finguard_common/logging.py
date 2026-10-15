"""Structured logging for FinGuard Python services."""

from __future__ import annotations

import logging
import os
import sys
from typing import Any

import structlog


def configure_logging(level: str | None = None) -> None:
    """Configure structlog + stdlib logging for Cloud Logging compatibility."""
    log_level = (level or os.getenv("LOG_LEVEL") or "INFO").upper()
    logging.basicConfig(
        format="%(message)s",
        stream=sys.stdout,
        level=log_level,
    )
    structlog.configure(
        processors=[
            structlog.contextvars.merge_contextvars,
            structlog.processors.add_log_level,
            structlog.processors.TimeStamper(fmt="iso", utc=True),
            _add_severity,
            structlog.processors.StackInfoRenderer(),
            structlog.processors.format_exc_info,
            structlog.processors.JSONRenderer(),
        ],
        wrapper_class=structlog.make_filtering_bound_logger(getattr(logging, log_level)),
        context_class=dict,
        logger_factory=structlog.PrintLoggerFactory(),
        cache_logger_on_first_use=True,
    )


def _add_severity(_: Any, __: str, event_dict: dict[str, Any]) -> dict[str, Any]:
    """Add a Cloud Logging-compatible `severity` field."""
    level = event_dict.get("level", "info")
    mapping = {
        "debug": "DEBUG",
        "info": "INFO",
        "warning": "WARNING",
        "error": "ERROR",
        "critical": "CRITICAL",
    }
    event_dict["severity"] = mapping.get(str(level).lower(), "INFO")
    return event_dict


def get_logger(name: str = "finguard") -> Any:
    """Return a structlog logger pre-configured with the service name."""
    return structlog.get_logger(name)
