-- alerts.sql
-- BigQuery DDL for analyst-facing fraud alerts. Populated by the
-- alert-worker Pub/Sub subscriber and updated by analyst actions.

CREATE TABLE IF NOT EXISTS `finguard.alerts`
(
  alert_id      STRING    NOT NULL,
  txn_id        STRING    NOT NULL,
  customer_id   STRING    NOT NULL,
  severity      STRING    NOT NULL, -- LOW / MEDIUM / HIGH / CRITICAL
  score         FLOAT64   NOT NULL,
  decision      STRING    NOT NULL, -- APPROVE / REVIEW / BLOCK
  status        STRING    NOT NULL DEFAULT 'OPEN', -- OPEN / IN_REVIEW / ESCALATED / CLOSED
  assignee      STRING,
  created_at    TIMESTAMP NOT NULL,
  updated_at    TIMESTAMP NOT NULL,
  closed_at     TIMESTAMP,
  closure_note  STRING,
  model_version STRING,
  features      JSON,
  shap_top      JSON                  -- [{feature, value, sign}]
)
PARTITION BY DATE(created_at)
CLUSTER BY status, severity, customer_id
OPTIONS (
  description = "FinGuard AI fraud alerts."
);
