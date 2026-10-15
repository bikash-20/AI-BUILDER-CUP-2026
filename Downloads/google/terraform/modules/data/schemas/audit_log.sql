-- audit_log.sql
-- BigQuery DDL for analyst/system audit log. Captures every analyst
-- action on a case, every SAR draft, and every config change.

CREATE TABLE IF NOT EXISTS `finguard.audit_log`
(
  event_id    STRING    NOT NULL,
  actor       STRING    NOT NULL,    -- user/service account that performed the action
  actor_type  STRING    NOT NULL,    -- ANALYST / SYSTEM / SERVICE
  action      STRING    NOT NULL,    -- ALERT_VIEW / CASE_OPEN / CASE_CLOSE / SAR_DRAFT / CONFIG_UPDATE
  resource    STRING,                -- target resource (alert_id, customer_id, ...)
  before      JSON,                  -- pre-state for updates
  after       JSON,                  -- post-state
  request_id  STRING,                -- correlation id from the originating request
  ip_address  STRING,
  user_agent  STRING,
  occurred_at TIMESTAMP NOT NULL
)
PARTITION BY DATE(occurred_at)
CLUSTER BY actor, action, resource
OPTIONS (
  description = "FinGuard AI audit log (immutable, 7-year retention via dataset config)."
);
