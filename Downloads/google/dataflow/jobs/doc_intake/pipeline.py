"""doc_intake pipeline.

Watches a GCS bucket for new objects, calls the document_service extract
endpoint, and writes results to BigQuery.
"""

from __future__ import annotations

import json
import sys
from pathlib import Path

import apache_beam as beam
from apache_beam.io import ReadAllFromText
from apache_beam.io.gcp.bigquery import WriteToBigQuery
from apache_beam.io.gcp.gcsio import parse_gcs_url

sys.path.insert(0, str(Path(__file__).resolve().parent))
from options import build_options  # noqa: E402
from transforms import ExtractViaDocAI, GCSFileToMessage  # noqa: E402


def discover_files(bucket: str) -> list[dict[str, object]]:
    """Discover object metadata for a bucket (synchronous helper for tests)."""
    from apache_beam.io.gcp import gcsio  # type: ignore

    gcs = gcsio.GcsIO()
    files: list[dict[str, object]] = []
    for path in gcs.list_prefix(bucket):
        m = parse_gcs_url(path)
        files.append({"bucket": m.bucket, "object": m.object})
    return files


def run(argv: list[str] | None = None) -> int:
    options, known = build_options(argv)
    p = beam.Pipeline(options=options)

    files = discover_files(known.input_bucket)

    (
        p
        | "Files" >> beam.Create(files)
        | "ToMessage" >> beam.ParDo(GCSFileToMessage())
        | "Extract" >> beam.ParDo(ExtractViaDocAI())
        | "ToJSON" >> beam.Map(json.dumps)
        | "WriteBQ" >> WriteToBigQuery(
            table=known.bq_table,
            schema={
                "fields": [
                    {"name": "gcs_uri", "type": "STRING", "mode": "REQUIRED"},
                    {"name": "size", "type": "INT64", "mode": "NULLABLE"},
                    {
                        "name": "extracted",
                        "type": "RECORD",
                        "mode": "NULLABLE",
                        "fields": [
                            {"name": "doc_type", "type": "STRING", "mode": "NULLABLE"},
                            {"name": "confidence", "type": "FLOAT64", "mode": "NULLABLE"},
                        ],
                    },
                ]
            },
            write_disposition=beam.io.BigQueryDisposition.WRITE_APPEND,
            create_disposition=beam.io.BigQueryDisposition.CREATE_IF_NEEDED,
        )
    )

    p.run().wait_until_finish()
    return 0


if __name__ == "__main__":
    sys.exit(run())
