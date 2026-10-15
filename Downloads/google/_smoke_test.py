"""Smoke test for the gemini_service."""
import sys
from pathlib import Path

# Insert the service directory and libs
SERVICE_DIR = Path(__file__).resolve().parent / "ml/services/gemini_service"
LIBS_DIR = Path(__file__).resolve().parent / "ml/libs"
sys.path.insert(0, str(SERVICE_DIR))
sys.path.insert(0, str(LIBS_DIR))

import os
os.environ["MOCK_MODE"] = "true"

from app import app
from fastapi.testclient import TestClient

print("App:", app.title)
c = TestClient(app)
r = c.get("/healthz")
print("status:", r.status_code, "data:", r.data)

r = c.post("/chat", json={"session_id": "test-1", "message": "Why was my card declined?"})
print("chat status:", r.status_code, "body:", r.json())

r = c.post("/summarize", json={"case_id": "C-1", "body": "hello " * 50})
print("summarize status:", r.status_code, "body:", r.json())
