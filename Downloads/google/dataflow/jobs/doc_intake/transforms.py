"""Beam transforms for the doc_intake job."""

from __future__ import annotations

import json
from typing import Any

import apache_beam as beam


class GCSFileToMessage(beam.DoFn):
    """Reads a GCS file's metadata and produces a message envelope."""

    def process(self, file_metadata: dict[str, Any]):
        yield {
            "gcs_uri": f"gs://{file_metadata['bucket']}/{file_metadata['object']}",
            "size": file_metadata.get("size", 0),
        }


class ExtractViaDocAI(beam.DoFn):
    """Calls the document_service HTTP endpoint to extract entities."""

    def process(self, element: dict[str, Any]):
        # Real implementation: requests.post(docai_url, json=element)
        # Mock returns a synthetic response.
        element["extracted"] = {
            "doc_type": "KYC",
            "entities": [{"type": "person_name", "value": "MOCK", "confidence": 0.99}],
            "confidence": 0.99,
        }
        yield element
