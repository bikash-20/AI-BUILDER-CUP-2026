"""Feature engineering pipeline (BigQuery + Dataflow)."""

from __future__ import annotations

import argparse
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from kfp import compiler, dsl  # type: ignore

from pipelines.components import (  # type: ignore
    feature_eng,
)


@dsl.pipeline(
    name="feature-engineering",
    description="Recompute customer feature MV in BigQuery.",
)
def feature_engineering_pipeline(
    project: str = "finguard-dev",
    dataset: str = "fin",
) -> None:
    feature_eng(project=project, dataset=dataset).set_display_name("Refresh feature MV")


def main() -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--project", required=True)
    p.add_argument("--compile-only", action="store_true")
    args = p.parse_args()
    if args.compile_only:
        out = Path(__file__).parent / "compiled" / "feature_engineering_pipeline.yaml"
        out.parent.mkdir(parents=True, exist_ok=True)
        compiler.Compiler().compile(feature_engineering_pipeline, str(out))
        print(f"compiled -> {out}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
