"""Document AI processor fine-tuning pipeline."""

from __future__ import annotations

import argparse
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from kfp import compiler, dsl  # type: ignore


@dsl.pipeline(
    name="docai-finetune",
    description="Fine-tune a Document AI processor on labeled KYC docs.",
)
def docai_finetune_pipeline(
    project: str = "finguard-dev",
    location: str = "us",
    processor_id: str = "REPLACE_ME",
    dataset_uri: str = "gs://finguard-documents/train.jsonl",
) -> None:
    from pipelines.components import (  # type: ignore
        data_validation,
    )

    valid = data_validation(project=project, dataset="fin", table="docai_labels")
    # Real implementation would call update_processor with the dataset URI.
    valid.set_display_name("Validate labeled dataset")


def main() -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--project", required=True)
    p.add_argument("--location", default="us")
    p.add_argument("--compile-only", action="store_true")
    args = p.parse_args()
    if args.compile_only:
        out = Path(__file__).parent / "compiled" / "docai_finetune_pipeline.yaml"
        out.parent.mkdir(parents=True, exist_ok=True)
        compiler.Compiler().compile(docai_finetune_pipeline, str(out))
        print(f"compiled -> {out}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
