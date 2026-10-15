"""Common utilities for training jobs."""

from __future__ import annotations

import hashlib
import json
import os
from typing import Any


def stable_run_id(prefix: str = "run") -> str:
    """Generate a stable run id from a seed and current time bucket."""
    seed = os.getenv("RUN_SEED", "0")
    bucket = os.getenv("RUN_TIME_BUCKET", "hourly")
    h = hashlib.sha256(f"{prefix}|{seed}|{bucket}".encode()).hexdigest()[:12]
    return f"{prefix}-{h}"


def write_json(path: str, data: Any) -> None:
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as f:
        json.dump(data, f, indent=2, default=str)


def read_json(path: str) -> Any:
    with open(path) as f:
        return json.load(f)
