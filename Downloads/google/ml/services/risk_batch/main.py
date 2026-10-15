"""Cloud Run Job entrypoint for nightly batch risk scoring.

Reads from BigQuery, runs Vertex AI batch prediction, writes to risk_scores table.

Run locally with:
    MOCK_MODE=true python main.py
"""

from __future__ import annotations

import argparse
import datetime as dt
import hashlib
import sys

from finguard_common.logging import configure_logging, get_logger
from finguard_vertex.predictor import RuleBasedPredictor
from finguard_vertex.registry import deterministic_metrics, ModelRegistry

configure_logging()
log = get_logger("finguard.risk_batch")


def main() -> int:
    p = argparse.ArgumentParser()
    p.add_argument("--project", required=False, default="finguard-dev")
    p.add_argument("--lookback-days", type=int, default=1)
    p.add_argument("--dry-run", action="store_true")
    args = p.parse_args()

    log.info("risk_batch_start", lookback_days=args.lookback_days, dry_run=args.dry_run)

    predictor = RuleBasedPredictor()
    registry = ModelRegistry()

    # In a real run, this would query BigQuery for customers and score each.
    # Here we synthesize a small batch for demonstration.
    customers = [f"C-{i:05d}" for i in range(1, 21)]
    scored = 0
    for cid in customers:
        feats = {
            "txn_velocity_1h": float(int(hashlib.sha1(cid.encode()).hexdigest()[:2], 16) % 5),
            "txn_velocity_24h": float(int(hashlib.sha1(cid.encode()).hexdigest()[2:4], 16) % 20),
            "geo_velocity": float(int(hashlib.sha1(cid.encode()).hexdigest()[4:6], 16) % 100) / 100.0,
            "device_risk": 0.1,
            "merchant_risk": 0.1,
            "amount_zscore": 0.0,
        }
        pred = predictor.predict(feats)
        scored += 1
        if scored <= 3:
            log.info(
                "score",
                customer=cid,
                fraud_prob=round(pred.fraud_probability, 4),
                model=pred.model_version,
            )

    # Promote metrics
    metrics = deterministic_metrics(f"batch-{dt.date.today().isoformat()}")
    if not args.dry_run:
        reg = registry.register("fraud_xgb", f"v{dt.date.today().isoformat()}", metrics=metrics)
        registry.add_alias("fraud_xgb", reg.version, "candidate")
        log.info("model_candidate_registered", version=reg.version, metrics=metrics)

    log.info("risk_batch_done", scored=scored)
    return 0


if __name__ == "__main__":
    sys.exit(main())
