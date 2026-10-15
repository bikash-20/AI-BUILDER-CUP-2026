"""Tests for finguard_docai.validators."""

from __future__ import annotations

from finguard_docai.validators import AMLValidator, KYCValidator


def test_kyc_valid() -> None:
    v = KYCValidator()
    doc = {
        "entities": [
            {"type": "person_name", "value": "JOHN DOE", "confidence": 0.99},
            {"type": "id_number", "value": "ABCDE1234F", "confidence": 0.99},
            {"type": "dob", "value": "1990-01-01", "confidence": 0.95},
        ]
    }
    r = v.validate(doc)
    assert r.valid is True


def test_kyc_invalid_id() -> None:
    v = KYCValidator()
    doc = {"entities": [{"type": "id_number", "value": "bad", "confidence": 0.99}]}
    r = v.validate(doc)
    assert r.valid is False
    assert "invalid_id_format" in r.issues


def test_kyc_empty() -> None:
    v = KYCValidator()
    r = v.validate({"entities": []})
    assert r.needs_review is True


def test_aml_invoice_valid() -> None:
    v = AMLValidator()
    doc = {
        "entities": [
            {"type": "invoice_number", "value": "INV1"},
            {"type": "total_amount", "value": "100.5"},
            {"type": "due_date", "value": "2026-12-01"},
        ]
    }
    r = v.validate_invoice(doc)
    assert r.valid is True
