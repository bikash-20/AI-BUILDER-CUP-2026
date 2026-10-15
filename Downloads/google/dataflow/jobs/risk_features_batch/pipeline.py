"""risk_features_batch pipeline.

Reads from `fin.transactions` and writes per-customer feature rows to
`fin.customer_features_mv`.
"""

from __future__ import annotations

import sys
from pathlib import Path

import apache_beam as beam
from apache_beam.io.gcp.bigquery import ReadFromBigQuery, WriteToBigQuery

sys.path.insert(0, str(Path(__file__).resolve().parent))
from options import build_options  # noqa: E402
from transforms import ComputeVelocity, SumFeatures  # noqa: E402


def run(argv: list[str] | None = None) -> int:
    options, known = build_options(argv)
    p = beam.Pipeline(options=options)

    schema = {
        "fields": [
            {"name": "customer_id", "type": "STRING", "mode": "REQUIRED"},
            {"name": "txn_velocity_24h", "type": "INT64", "mode": "NULLABLE"},
            {"name": "amount_sum_24h", "type": "FLOAT64", "mode": "NULLABLE"},
        ]
    }

    (
        p
        | "ReadTransactions" >> ReadFromBigQuery(table=known.bq_source)
        | "ComputeVelocity" >> beam.ParDo(ComputeVelocity())
        | "SumPerCustomer"
            >> beam.CombinePerKey(SumFeatures())
        | "Format" >> beam.Map(lambda kv: {"customer_id": kv[0], **kv[1]})
        | "Write" >> WriteToBigQuery(
            table=known.bq_dest,
            schema=schema,
            write_disposition=beam.io.BigQueryDisposition.WRITE_TRUNCATE,
            create_disposition=beam.io.BigQueryDisposition.CREATE_IF_NEEDED,
        )
    )

    p.run().wait_until_finish()
    return 0


if __name__ == "__main__":
    sys.exit(run())
