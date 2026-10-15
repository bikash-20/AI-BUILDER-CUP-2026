"""Tests for risk_features_batch transforms."""

from __future__ import annotations

import os
import sys
import unittest

import apache_beam as beam
from apache_beam.testing.test_pipeline import TestPipeline
from apache_beam.testing.util import assert_that, equal_to

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))
from transforms import ComputeVelocity, SumFeatures  # noqa: E402


class TestRiskBatchTransforms(unittest.TestCase):
    def test_compute_velocity(self) -> None:
        with TestPipeline() as p:
            out = (
                p
                | beam.Create([{"customer_id": "C-1", "amount": 100.0}])
                | beam.ParDo(ComputeVelocity())
            )
            assert_that(
                out,
                equal_to([{"customer_id": "C-1", "txn_velocity_24h": 1, "amount_sum_24h": 100.0}]),
            )

    def test_sum_features_combine(self) -> None:
        with TestPipeline() as p:
            out = (
                p
                | beam.Create(
                    [
                        ("C-1", {"txn_velocity_24h": 1, "amount_sum_24h": 100.0}),
                        ("C-1", {"txn_velocity_24h": 1, "amount_sum_24h": 200.0}),
                    ]
                )
                | beam.CombinePerKey(SumFeatures())
            )
            assert_that(
                out,
                equal_to([("C-1", {"txn_velocity_24h": 2, "amount_sum_24h": 300.0})]),
            )


if __name__ == "__main__":
    unittest.main()
