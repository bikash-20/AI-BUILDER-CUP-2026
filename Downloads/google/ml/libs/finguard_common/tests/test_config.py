"""Tests for finguard_common.config."""

from __future__ import annotations

import os

import pytest

from finguard_common import config


@pytest.fixture(autouse=True)
def _reset(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("GCP_PROJECT_ID", "test-project")
    config.reset_cache()
    yield  # type: ignore[misc]
    config.reset_cache()


def test_settings_loaded() -> None:
    s = config.get_settings()
    assert s.project_id == "test-project"
    assert s.region == "asia-south1"


def test_is_mock_mode_default(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("MOCK_MODE", raising=False)
    config.reset_cache()
    assert config.is_mock_mode() is True


def test_is_mock_mode_off(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("MOCK_MODE", "false")
    config.reset_cache()
    assert config.is_mock_mode() is False


def test_feature_flags(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("FEATURE_FLAGS", " a=on , b = off , c")
    config.reset_cache()
    s = config.get_settings()
    assert s.feature_flags.get("a") is True
    assert s.feature_flags.get("b") is False
    assert s.feature_flags.get("c") is True
