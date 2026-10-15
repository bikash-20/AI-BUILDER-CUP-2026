"""Tests for compliance_llm FastAPI app."""

from __future__ import annotations

import os

os.environ.setdefault("MOCK_MODE", "true")

from fastapi.testclient import TestClient

from app import app  # noqa: E402

client = TestClient(app)


def test_healthz() -> None:
    r = client.get("/healthz")
    assert r.status_code == 200


def test_draft_sar() -> None:
    r = client.post(
        "/draft_sar",
        json={
            "case_id": "C-1",
            "transactions": [{"amount": 1500.0}],
            "notes": "high velocity",
        },
    )
    assert r.status_code == 200
    body = r.json()
    assert body["filing_type"] == "SAR"
    assert body["case_id"] == "C-1"
    assert "draft_id" in body


def test_draft_ctr() -> None:
    r = client.post("/draft_ctr", json={"case_id": "C-1", "transactions": []})
    assert r.status_code == 200
    assert r.json()["filing_type"] == "CTR"
