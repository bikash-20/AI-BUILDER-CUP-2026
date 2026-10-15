"""Conftest for the gemini_service tests."""

from __future__ import annotations

import os
import sys
from pathlib import Path

# Insert the service directory so `from app import app` works.
SERVICE_DIR = str(Path(__file__).resolve().parent.parent)
if SERVICE_DIR not in sys.path:
    sys.path.insert(0, SERVICE_DIR)

# Insert ml/libs so `import finguard_*` works.
LIBS_DIR = str(Path(__file__).resolve().parents[3] / "libs")
if LIBS_DIR not in sys.path:
    sys.path.insert(0, LIBS_DIR)

os.environ.setdefault("MOCK_MODE", "true")
