"""Unit tests for txn_enrichment transforms."""

from __future__ import annotations

import json
import os
import sys
import unittest

import apache_beam as beam
from apache_beam.testing.test_pipeline import TestPipeline
from apache_beam.testing.util import assert_that, equal_to

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))
from transforms import (  # noqa: E402
    EnrichWithFeatures,
    ParseJson,
    PublishToScoredTopic,
)


class TestTransforms(unittest.TestCase):
    def test_parse_json(self) -> None:
        with TestPipeline() as p:
            out = (
                p
                | beam.Create([b'{"a":1}', b"not-json"])
                | beam.ParDo(ParseJson()).with_outputs("bad", main="good")
            )
            assert_that(out.good, equal_to([{"a": 1}]), label="good")
            assert_that(out.bad, equal_to([b"not-json"]), label="bad")

    def test_enrich_with_features(self) -> None:
        with TestPipeline() as p:
            out = (
                p
                | beam.Create([{"txn_id": "T-1", "customer_id": "C-1"}])
                | beam.ParDo(EnrichWithFeatures())
            )
            assert_that(out, equal_to([{"txn_id": "T-1", "customer_id": "C-1", "_features": {"txn_velocity_1h": 0.0, "geo_velocity": 0.0, "device_risk": 0.0}}]))

    def test_publish_to_scored(self) -> None:
        with TestPipeline() as p:
            out = (
                p
                | beam.Create([{"a": 1}])
                | beam.ParDo(PublishToScoredTopic())
            )
            assert_that(out, equal_to([json.dumps({"a": 1}).encode("utf-8")]))


if __name__ == "__main__":
    unittest.main()
