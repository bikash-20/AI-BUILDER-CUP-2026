"""Tests for document_service FastAPI app."""

from __future__ import annotations

import os

os.environ.setdefault("MOCK_MODE", "true")

from fastapi.testclient import TestClient

from app import app  # noqa: E402

client = TestClient(app)


def test_healthz() -> None:
    r = client.get("/healthz")
    assert r.status_code == 200
    assert r.json() == {"status": "ok"}


def test_extract_mock() -> None:
    r = client.post(
        "/extract",
        json={"gcs_uri": "gs://bucket/file.pdf", "doc_type": "KYC"},
    )
    assert r.status_code == 200
    body = r.json()
    assert body["doc_type"] == "KYC"
    assert any(e["type"] == "person_name" for e in body["entities"])


def test_extract_invalid_uri() -> None:
    r = client.post("/extract", json={"gcs_uri": "http://x", "doc_type": "KYC"})
    assert r.status_code == 422


def test_validate_kyc() -> None:
    r = client.post(
        "/validate",
        json={
            "doc_type": "KYC",
            "entities": [
                {"type": "person_name", "value": "JOHN DOE", "confidence": 0.99},
                {"type": "id_number", "value": "ABCDE1234F", "confidence": 0.99},
                {"type": "dob", "value": "1990-01-01", "confidence": 0.95},
            ],
        },
    )
    assert r.status_code == 200
    body = r.json()
    assert body["valid"] is True
