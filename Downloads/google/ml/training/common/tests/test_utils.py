"""Tests for training.common.utils."""

from __future__ import annotations

import os
import sys
import tempfile

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))
import utils  # noqa: E402


def test_stable_run_id() -> None:
    rid = utils.stable_run_id("train")
    assert rid.startswith("train-")


def test_write_read_json(tmp_path) -> None:
    p = tmp_path / "x.json"
    utils.write_json(str(p), {"a": 1})
    assert utils.read_json(str(p)) == {"a": 1}
