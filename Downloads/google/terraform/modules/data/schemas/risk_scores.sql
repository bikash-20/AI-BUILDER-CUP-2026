-- risk_scores.sql
-- BigQuery DDL for customer-level risk scores over time.
-- Used by the risk-service for /v1/risk/{customer_id} reads and by
-- the risk-features-batch Dataflow job for nightly recomputation.

CREATE TABLE IF NOT EXISTS `finguard.risk_scores`
(
  customer_id     STRING    NOT NULL,
  score_id        STRING    NOT NULL,
  composite_score FLOAT64   NOT NULL, -- 0..1000
  band            STRING    NOT NULL, -- LOW / MEDIUM / HIGH / CRITICAL
  drivers         JSON,                -- top SHAP features
  computed_at     TIMESTAMP NOT NULL,
  model_version   STRING    NOT NULL,
  horizon_days    INT64,               -- prediction window
  source          STRING               -- BATCH / ONLINE / MANUAL
)
PARTITION BY DATE(computed_at)
CLUSTER BY customer_id, band
OPTIONS (
  description = "FinGuard AI customer risk score history."
);
