-- transactions.sql
-- BigQuery DDL for the FinGuard AI transactions table.
-- Partitioned by ingestion time (event_time), clustered by customer_id
-- for efficient point lookups during feature lookup and ad-hoc analytics.

CREATE TABLE IF NOT EXISTS `finguard.transactions`
(
  txn_id        STRING     NOT NULL,
  customer_id   STRING     NOT NULL,
  account_id    STRING,
  amount        NUMERIC    NOT NULL,
  currency      STRING     NOT NULL,
  merchant_id   STRING,
  merchant_name STRING,
  mcc           STRING,
  device_id     STRING,
  ip_address    STRING,
  geo_country   STRING,
  geo_city      STRING,
  event_time    TIMESTAMP  NOT NULL,
  ingested_at   TIMESTAMP  NOT NULL DEFAULT CURRENT_TIMESTAMP(),
  source        STRING,                 -- WEB / MOBILE / POS / API
  status        STRING,                 -- APPROVED / DECLINED / PENDING
  risk_score    FLOAT64,                -- 0..1, latest
  decision      STRING,                 -- APPROVE / REVIEW / BLOCK
  model_version STRING,
  -- Flexible carrier for arbitrary features / enrichment data.
  features      JSON
)
PARTITION BY DATE(event_time)
CLUSTER BY customer_id, merchant_id, decision
OPTIONS (
  description = "FinGuard AI transactions (per the PRD data model).",
  labels      = [("env", "ENV_PLACEHOLDER")]
);
