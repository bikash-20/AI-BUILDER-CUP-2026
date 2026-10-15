"""Tests for finguard_common.clients (mock mode)."""

from __future__ import annotations

import pytest

from finguard_common import clients, config


@pytest.fixture(autouse=True)
def _mock(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("MOCK_MODE", "true")
    config.reset_cache()


def test_bigquery_client() -> None:
    c = clients.bigquery_client()
    assert hasattr(c, "query")


def test_documentai_client() -> None:
    c = clients.documentai_client()
    assert hasattr(c, "process_document")


def test_dlp_client() -> None:
    c = clients.dlp_client()
    assert hasattr(c, "deidentify_content")


def test_pubsub_publisher() -> None:
    p = clients.pubsub_publisher()
    assert p.publish("t", b"x") == "mock-message-id"


def test_storage_client() -> None:
    s = clients.storage_client()
    assert s.bucket("b").blob("k") is not None


def test_secret_manager_client() -> None:
    s = clients.secret_manager_client()
    assert s.access_secret_version(name="x") is not None
