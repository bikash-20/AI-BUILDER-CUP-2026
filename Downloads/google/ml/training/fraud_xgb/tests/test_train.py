"""Tests for fraud_xgb training."""

from __future__ import annotations

import os
import subprocess
import sys

ROOT = os.path.join(os.path.dirname(__file__), "..")


def test_train_mock_runs() -> None:
    env = os.environ.copy()
    env["MOCK_MODE"] = "true"
    p = subprocess.run(
        [sys.executable, "train.py", "--mock", "--model-dir", "/tmp/finguard-test-model"],
        cwd=ROOT,
        env=env,
        capture_output=True,
        text=True,
    )
    assert p.returncode == 0, p.stderr
