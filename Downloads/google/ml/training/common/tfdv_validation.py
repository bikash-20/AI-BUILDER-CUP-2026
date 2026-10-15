"""TFDV-based data validation for training and serving datasets.

Used as a guard at every pipeline boundary.
"""

from __future__ import annotations

import json
import os
import sys
from typing import Any


def validate_dataset_stats(stats_path: str) -> dict[str, Any]:
    """Validate a TFDV stats file against an expected schema.

    Real implementation would use:
        from tensorflow_data_validation import validate_statistics
    For now, we return a simple structural check.
    """
    if not os.path.exists(stats_path):
        return {"ok": False, "error": "stats file not found", "path": stats_path}
    try:
        with open(stats_path) as f:
            stats = json.load(f)
        if "datasets" not in stats:
            return {"ok": False, "error": "missing 'datasets' key"}
        return {"ok": True, "datasets": len(stats["datasets"])}
    except json.JSONDecodeError as e:
        return {"ok": False, "error": f"invalid json: {e}"}


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("usage: tfdv_validation.py <stats.json>")
        sys.exit(1)
    print(json.dumps(validate_dataset_stats(sys.argv[1]), indent=2))
