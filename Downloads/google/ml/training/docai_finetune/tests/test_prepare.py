"""Tests for docai_finetune.prepare_dataset."""

from __future__ import annotations

import os
import sys
import tempfile

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))
import prepare_dataset  # noqa: E402


def test_write_dataset(tmp_path) -> None:
    out = tmp_path / "dataset.jsonl"
    n = prepare_dataset.write_dataset(str(out))
    assert n >= 2
    lines = out.read_text().strip().splitlines()
    assert len(lines) == n
