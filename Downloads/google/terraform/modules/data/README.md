# data

Creates the FinGuard AI data plane:

- **BigQuery** dataset for transactional and ML feature data. Tables
  (transactions, risk_scores, documents, alerts, audit_log) are created
  via SQL files in `schemas/` and applied with `google_bigquery_table`
  resources here.
- **Pub/Sub** topics for event streaming (transaction events, scoring
  requests, document intake, alert fanout).
- **GCS** buckets for raw document uploads, training artifacts, and
  Dataflow staging.
- All resources are encrypted with CMEK keys from the `security` module.
