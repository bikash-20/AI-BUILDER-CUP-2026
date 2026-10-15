"""Continuous training pipeline triggered by drift signals or schedule."""

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
    name="fraud-continuous-training",
    description="Triggered CT pipeline for fraud model.",
)
def fraud_continuous_training(
    project: str = "finguard-dev",
    region: str = "asia-south1",
    trigger: str = "schedule",
    feature: str = "txn_velocity_24h",
    accuracy_threshold: float = 0.85,
) -> None:
    valid_task = data_validation(project=project, dataset="fin", table="training_data")
    feat_task = feature_eng(project=project, dataset="fin").after(valid_task)
    train_task = train(project=project, dataset="fin", table="training_data").after(feat_task)
    eval_task = evaluate(model_dir=train_task.outputs["model_dir"]).after(train_task)

    with dsl.Condition(eval_task.outputs["metrics"] >= accuracy_threshold, name="promote"):
        reg_task = register(
            project=project, region=region, metrics=eval_task.outputs["metrics"]
        ).after(eval_task)
        deploy(
            project=project,
            region=region,
            model=reg_task.outputs["model"],
            traffic_percent=10,
        ).after(reg_task)


def main() -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--project", required=True)
    p.add_argument("--region", default="asia-south1")
    p.add_argument("--trigger", default="schedule")
    p.add_argument("--compile-only", action="store_true")
    args = p.parse_args()

    if args.compile_only:
        out = Path(__file__).parent / "compiled" / "fraud_continuous_training.yaml"
        out.parent.mkdir(parents=True, exist_ok=True)
        compiler.Compiler().compile(fraud_continuous_training, str(out))
        print(f"compiled -> {out}")
    else:
        from kfp.client import Client  # type: ignore

        client = Client()
        client.create_run_from_pipeline_func(
            fraud_continuous_training,
            arguments={
                "project": args.project,
                "region": args.region,
                "trigger": args.trigger,
            },
        )
    return 0


if __name__ == "__main__":
    sys.exit(main())
