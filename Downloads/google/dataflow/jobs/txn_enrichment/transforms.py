"""Beam transforms for the txn_enrichment job.

Implements the FR-1 ingestion + enrichment flow.
"""

from __future__ import annotations

import json
from typing import Any

import apache_beam as beam
from apache_beam import pvalue


class ParseJson(beam.DoFn):
    def process(self, element: bytes):
        try:
            data = json.loads(element.decode("utf-8"))
            yield data
        except Exception:
            yield pvalue.TaggedOutput("bad", element)


class EnrichWithFeatures(beam.DoFn):
    """Look up BigQuery features for the customer and attach to the txn."""

    def process(self, txn: dict[str, Any]):
        # Real implementation: BigQuery row lookup via side input or client.
        features = {"txn_velocity_1h": 0.0, "geo_velocity": 0.0, "device_risk": 0.0}
        yield {**txn, "_features": features}


class PublishToScoredTopic(beam.DoFn):
    def process(self, element: dict[str, Any]):
        yield json.dumps(element).encode("utf-8")
