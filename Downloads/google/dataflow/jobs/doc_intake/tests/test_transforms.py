"""Tests for doc_intake transforms."""

from __future__ import annotations

import os
import sys
import unittest

import apache_beam as beam
from apache_beam.testing.test_pipeline import TestPipeline
from apache_beam.testing.util import assert_that, equal_to

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))
from transforms import ExtractViaDocAI, GCSFileToMessage  # noqa: E402


class TestDocIntakeTransforms(unittest.TestCase):
    def test_gcs_file_to_message(self) -> None:
        with TestPipeline() as p:
            out = (
                p
                | beam.Create([{"bucket": "b", "object": "o.pdf", "size": 100}])
                | beam.ParDo(GCSFileToMessage())
            )
            assert_that(out, equal_to([{"gcs_uri": "gs://b/o.pdf", "size": 100}]))

    def test_extract_via_docai(self) -> None:
        with TestPipeline() as p:
            out = (
                p
                | beam.Create([{"gcs_uri": "gs://b/o.pdf"}])
                | beam.ParDo(ExtractViaDocAI())
            )
            assert_that(
                out,
                equal_to(
                    [
                        {
                            "gcs_uri": "gs://b/o.pdf",
                            "extracted": {
                                "doc_type": "KYC",
                                "entities": [
                                    {"type": "person_name", "value": "MOCK", "confidence": 0.99}
                                ],
                                "confidence": 0.99,
                            },
                        }
                    ]
                ),
            )


if __name__ == "__main__":
    unittest.main()
