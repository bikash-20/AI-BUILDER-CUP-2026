"""Entrypoint for the document_service Cloud Run service.

Run with:
    MOCK_MODE=true uvicorn app:app --host 0.0.0.0 --port 8080
"""

from __future__ import annotations

from finguard_common.logging import configure_logging
from finguard_docai import MockDocumentExtractor
from finguard_docai.extractor import DocumentAIExtractor
from finguard_docai.validators import KYCValidator

configure_logging()

from app import app  # noqa: E402  (re-export)

__all__ = ["app"]


def _get_extractor():  # pragma: no cover
    from finguard_common.config import is_mock_mode

    if is_mock_mode():
        return MockDocumentExtractor()
    return DocumentAIExtractor()


def _get_validator():  # pragma: no cover
    return KYCValidator()
