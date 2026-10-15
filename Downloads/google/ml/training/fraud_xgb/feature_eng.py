"""BigQuery feature engineering for fraud detection.

Produces the `fin.customer_features_mv` materialized view used by fraud-service
and Vertex AI Feature Store.
"""

from __future__ import annotations

# Standard SQL with CREATE MATERIALIZED VIEW (BigQuery).
# The query is intentionally idempotent: safe to re-run.

DDL = """
CREATE OR REPLACE MATERIALIZED VIEW `{{project}}.fin.customer_features_mv` AS
SELECT
  customer_id,
  COUNTIF(DATE(txn_ts) >= DATE_SUB(CURRENT_DATE(), INTERVAL 1 HOUR))
    AS txn_velocity_1h,
  COUNTIF(DATE(txn_ts) >= DATE_SUB(CURRENT_DATE(), INTERVAL 1 DAY))
    AS txn_velocity_24h,
  -- geo velocity: max distance traveled in 1 hour
  MAX(geo_distance_km) AS max_geo_velocity_km,
  -- device risk: # distinct devices / total txns
  SAFE_DIVIDE(COUNT(DISTINCT device_id), COUNT(*)) AS device_diversity,
  -- merchant risk: avg risk for merchant
  AVG(merchant_risk) AS avg_merchant_risk,
  -- amount z-score vs 30-day average
  SAFE_DIVIDE(amount - avg_amount_30d, stddev_amount_30d) AS amount_zscore,
  -- behavioral: % txns that are foreign geo
  SAFE_DIVIDE(SUM(is_foreign_geo), COUNT(*)) AS foreign_geo_rate
FROM `{{project}}.fin.transactions`
WHERE txn_ts >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 90 DAY)
GROUP BY customer_id
"""


def render(project: str) -> str:
    return DDL.replace("{{project}}", project)
