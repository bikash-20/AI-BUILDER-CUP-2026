"""Tests for risk_batch main."""

from __future__ import annotations

import os
import sys

os.environ.setdefault("MOCK_MODE", "true")
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

import main as risk_main  # noqa: E402


def test_main_runs(capsys) -> None:
    rc = risk_main.main()
    assert rc == 0
