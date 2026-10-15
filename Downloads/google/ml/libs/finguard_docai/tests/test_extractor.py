"""Tests for finguard_docai.extractor."""

from __future__ import annotations

from finguard_docai.extractor import MockDocumentExtractor


def test_mock_kyc_entities() -> None:
    e = MockDocumentExtractor().extract("gs://x/y.pdf", "KYC")
    types = {ent["type"] for ent in e.entities}
    assert "person_name" in types
    assert "id_number" in types


def test_mock_invoice_entities() -> None:
    e = MockDocumentExtractor().extract("gs://x/y.pdf", "INVOICE")
    types = {ent["type"] for ent in e.entities}
    assert "total_amount" in types
