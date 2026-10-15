"""Prepare a Document AI fine-tuning dataset from labeled GCS docs.

For real use, replace `mock_documents` with a GCS listing + label file.
"""

from __future__ import annotations

import json
import os
from typing import Any

from finguard_common.logging import configure_logging, get_logger

configure_logging()
log = get_logger("finguard.training.docai_finetune.prepare")


def mock_documents() -> list[dict[str, Any]]:
    return [
        {
            "gcs_uri": "gs://finguard-documents/train/kyc-001.pdf",
            "doc_type": "KYC",
            "labels": [
                {"type": "person_name", "value": "JOHN DOE", "page": 1},
                {"type": "id_number", "value": "ABCDE1234F", "page": 1},
            ],
        },
        {
            "gcs_uri": "gs://finguard-documents/train/kyc-002.pdf",
            "doc_type": "KYC",
            "labels": [
                {"type": "person_name", "value": "JANE SMITH", "page": 1},
                {"type": "id_number", "value": "PQRST5678K", "page": 1},
            ],
        },
    ]


def write_dataset(out_path: str) -> int:
    os.makedirs(os.path.dirname(out_path), exist_ok=True)
    docs = mock_documents()
    with open(out_path, "w") as f:
        for d in docs:
            f.write(json.dumps(d) + "\n")
    log.info("dataset_written", count=len(docs), path=out_path)
    return len(docs)


if __name__ == "__main__":
    import sys

    out = sys.argv[1] if len(sys.argv) > 1 else "/tmp/finguard-docai-train.jsonl"
    n = write_dataset(out)
    print(json.dumps({"written": n, "path": out}))
