"""Tests for fraud_autoencoder training."""

from __future__ import annotations

import os
import subprocess
import sys


def test_autoencoder_train_mock() -> None:
    env = os.environ.copy()
    env["MOCK_MODE"] = "true"
    p = subprocess.run(
        [sys.executable, "train.py", "--mock", "--model-dir", "/tmp/finguard-test-ae", "--epochs", "1"],
        cwd=os.path.join(os.path.dirname(__file__), ".."),
        env=env,
        capture_output=True,
        text=True,
    )
    assert p.returncode == 0, p.stderr
