"""Beam transforms for the risk_features_batch job."""

from __future__ import annotations

from typing import Any

import apache_beam as beam


class ComputeVelocity(beam.DoFn):
    """Aggregate per-customer transaction counts into features."""

    def process(self, row: dict[str, Any]):
        yield {
            "customer_id": row["customer_id"],
            "txn_velocity_24h": 1,
            "amount_sum_24h": float(row.get("amount", 0.0)),
        }


class SumFeatures(beam.CombineFn):
    def create_accumulator(self) -> dict[str, float]:
        return {"txn_velocity_24h": 0, "amount_sum_24h": 0.0}

    def add_input(self, acc: dict[str, float], value: dict[str, float]) -> dict[str, float]:
        acc["txn_velocity_24h"] += value["txn_velocity_24h"]
        acc["amount_sum_24h"] += value["amount_sum_24h"]
        return acc

    def merge_accumulators(self, accs: list[dict[str, float]]) -> dict[str, float]:
        out = self.create_accumulator()
        for a in accs:
            out["txn_velocity_24h"] += a["txn_velocity_24h"]
            out["amount_sum_24h"] += a["amount_sum_24h"]
        return out

    def extract_output(self, acc: dict[str, float]) -> dict[str, float]:
        return acc
