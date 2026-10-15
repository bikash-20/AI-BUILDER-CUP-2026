# logging

Configures platform-wide log routing:

- **Log sinks** that forward Cloud Audit Logs and VPC flow logs to a
  BigQuery dataset for long-term analytics and to Pub/Sub for streaming
  alerting.
- A **BigQuery dataset** (`finguard_<env>_logs`) that holds log records
  using a default table expiration of 365 days.
- A **log-based metric** for tracking fraud-decision outcomes, which the
  monitoring module uses to power an SLO dashboard.
