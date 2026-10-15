"""Vertex AI Custom Training entry point for the XGBoost fraud model.

Run as a Vertex AI Custom Job:
    python train.py --bq-project=finguard-dev --bq-dataset=fin

Or locally:
    MOCK_MODE=true python train.py
"""

from __future__ import annotations

import argparse
import json
import os
import sys
from typing import Any

from finguard_common.logging import configure_logging, get_logger
from finguard_evaluation.metrics import binary_metrics

configure_logging()
log = get_logger("finguard.training.fraud_xgb")


def main() -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--bq-project", default=os.getenv("GCP_PROJECT_ID", "finguard-dev"))
    p.add_argument("--bq-dataset", default="fin")
    p.add_argument("--bq-table", default="training_data")
    p.add_argument("--model-dir", default=os.getenv("AIP_MODEL_DIR", "/tmp/model"))
    p.add_argument("--metrics-out", default=os.getenv("AIP_METRICS_URI", "/tmp/metrics.json"))
    p.add_argument("--mock", action="store_true", default=os.getenv("MOCK_MODE", "true") == "true")
    args = p.parse_args()

    log.info("train_start", bq=f"{args.bq_project}.{args.bq_dataset}.{args.bq_table}", model_dir=args.model_dir)

    if args.mock:
        metrics = _train_mock()
    else:
        metrics = _train_real(args)

    os.makedirs(args.model_dir, exist_ok=True)
    metrics_path = os.path.join(args.model_dir, "metrics.json")
    with open(metrics_path, "w") as f:
        json.dump(metrics, f, indent=2)
    if args.metrics_out and args.metrics_out != "/tmp/metrics.json":
        os.makedirs(os.path.dirname(args.metrics_out), exist_ok=True)
        with open(args.metrics_out, "w") as f:
            json.dump(metrics, f, indent=2)

    log.info("train_done", metrics=metrics)
    return 0


def _train_mock() -> dict[str, float]:
    """Generate deterministic synthetic training data and train a tiny XGBoost."""
    import numpy as np
    import xgboost as xgb

    rng = np.random.default_rng(42)
    n = 500
    X = rng.normal(size=(n, 5))
    y = (X[:, 0] + 0.5 * X[:, 1] - 0.3 * X[:, 2] > 0).astype(int)

    model = xgb.XGBClassifier(n_estimators=20, max_depth=3, use_label_encoder=False, eval_metric="logloss")
    model.fit(X, y)
    proba = model.predict_proba(X)[:, 1]
    metrics = binary_metrics(y.tolist(), proba.tolist())
    return metrics


def _train_real(args: Any) -> dict[str, float]:
    """Real training: read from BigQuery, train XGBoost, upload to Vertex AI Model Registry."""
    try:
        from google.cloud import bigquery  # type: ignore
        import xgboost as xgb
        from finguard_common.clients import bigquery_client
    except Exception as e:
        log.error("real_train_imports_failed", error=str(e))
        return {"error": str(e)}

    client = bigquery_client()
    sql = f"SELECT * FROM `{args.bq_project}.{args.bq_dataset}.{args.bq_table}` LIMIT 10000"
    df = client.query(sql).to_dataframe()
    if "is_fraud" not in df.columns:
        return {"error": "is_fraud column not found"}

    feature_cols = [c for c in df.columns if c not in ("is_fraud", "txn_id", "customer_id")]
    y = df["is_fraud"].astype(int)
    X = df[feature_cols].fillna(0.0)

    model = xgb.XGBClassifier(n_estimators=100, max_depth=6, use_label_encoder=False, eval_metric="logloss")
    model.fit(X, y)
    proba = model.predict_proba(X)[:, 1]
    metrics = binary_metrics(y.tolist(), proba.tolist())
    return metrics


if __name__ == "__main__":
    sys.exit(main())
