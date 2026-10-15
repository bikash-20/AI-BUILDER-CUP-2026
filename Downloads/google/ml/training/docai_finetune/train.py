"""Vertex AI Custom Training entry point for Document AI processor fine-tune.

In production, this would call `update_processor` with a labeled dataset
uploaded to a GCS bucket; here we scaffold the entry point and write a
manifest that documents the expected API calls.
"""

from __future__ import annotations

import argparse
import json
import os
import sys

from finguard_common.logging import configure_logging, get_logger

configure_logging()
log = get_logger("finguard.training.docai_finetune")


def main() -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--project", default=os.getenv("GCP_PROJECT_ID", "finguard-dev"))
    p.add_argument("--location", default="us")
    p.add_argument("--processor-id", required=False)
    p.add_argument("--dataset", default="/tmp/finguard-docai-train.jsonl")
    p.add_argument("--mock", action="store_true", default=os.getenv("MOCK_MODE", "true") == "true")
    args = p.parse_args()

    log.info("docai_finetune_start", processor_id=args.processor_id, mock=args.mock)

    if not os.path.exists(args.dataset):
        log.warning("dataset_missing", path=args.dataset)

    manifest = {
        "operation": "update_processor",
        "project": args.project,
        "location": args.location,
        "processor_id": args.processor_id,
        "dataset_uri": args.dataset,
        "mock": args.mock,
    }
    out_path = os.path.join(os.getenv("AIP_MODEL_DIR", "/tmp"), "docai-manifest.json")
    with open(out_path, "w") as f:
        json.dump(manifest, f, indent=2)
    log.info("docai_finetune_done", manifest=out_path)
    return 0


if __name__ == "__main__":
    sys.exit(main())
