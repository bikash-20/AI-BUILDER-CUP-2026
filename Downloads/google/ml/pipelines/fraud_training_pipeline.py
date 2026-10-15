"""Vertex AI Pipeline: end-to-end fraud model training.

Steps:
  1. Data validation (TFDV)
  2. Feature engineering
  3. XGBoost training
  4. Evaluation
  5. Conditional model registration
  6. Canary deployment
"""

from __future__ import annotations

import argparse
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from kfp import compiler, dsl  # type: ignore

from pipelines.components import (  # type: ignore
    data_validation,
    deploy,
    evaluate,
    feature_eng,
    register,
    train,
)


@dsl.pipeline(
    name="fraud-training-pipeline",
    description="Train and register a new fraud detection model.",
)
def fraud_training_pipeline(
    project: str = "finguard-dev",
    region: str = "asia-south1",
    bq_dataset: str = "fin",
    bq_table: str = "training_data",
    accuracy_threshold: float = 0.85,
) -> None:
    """End-to-end fraud training pipeline."""
    valid_task = data_validation(
        project=project, dataset=bq_dataset, table=bq_table
    ).set_display_name("Data Validation")

    feat_task = feature_eng(
        project=project, dataset=bq_dataset
    ).after(valid_task).set_display_name("Feature Engineering")

    train_task = train(
        project=project,
        dataset=bq_dataset,
        table=bq_table,
    ).after(feat_task).set_display_name("Train XGBoost")

    eval_task = evaluate(
        model_dir=train_task.outputs["model_dir"],
    ).after(train_task).set_display_name("Evaluate")

    with dsl.Condition(eval_task.outputs["metrics"] >= accuracy_threshold, name="promote"):
        reg_task = register(
            project=project,
            region=region,
            metrics=eval_task.outputs["metrics"],
        ).after(eval_task).set_display_name("Register Model")
        deploy(
            project=project,
            region=region,
            model=reg_task.outputs["model"],
            traffic_percent=10,
        ).after(reg_task).set_display_name("Canary Deploy")


def main() -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--project", required=True)
    p.add_argument("--region", default="asia-south1")
    p.add_argument("--compile-only", action="store_true")
    args = p.parse_args()

    if args.compile_only:
        out = Path(__file__).parent / "compiled" / "fraud_training_pipeline.yaml"
        out.parent.mkdir(parents=True, exist_ok=True)
        compiler.Compiler().compile(fraud_training_pipeline, str(out))
        print(f"compiled -> {out}")
    else:
        from kfp.client import Client  # type: ignore

        client = Client()
        client.create_run_from_pipeline_func(
            fraud_training_pipeline,
            arguments={"project": args.project, "region": args.region},
        )
    return 0


if __name__ == "__main__":
    sys.exit(main())
