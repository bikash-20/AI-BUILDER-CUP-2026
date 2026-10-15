"""Tests for gemini_service FastAPI app."""

from __future__ import annotations

import os

os.environ.setdefault("MOCK_MODE", "true")

from fastapi.testclient import TestClient

from app import app  # noqa: E402

client = TestClient(app)


def test_healthz() -> None:
    r = client.get("/healthz")
    assert r.status_code == 200


def test_chat_mock() -> None:
    r = client.post("/chat", json={"message": "Why was my card declined?"})
    assert r.status_code == 200
    body = r.json()
    assert body["reply"]
    assert "session_id" in body


def test_chat_empty() -> None:
    r = client.post("/chat", json={"message": ""})
    assert r.status_code == 400


def test_summarize() -> None:
    r = client.post("/summarize", json={"case_id": "C-1", "body": "hello " * 200})
    assert r.status_code == 200
    body = r.json()
    assert body["case_id"] == "C-1"
