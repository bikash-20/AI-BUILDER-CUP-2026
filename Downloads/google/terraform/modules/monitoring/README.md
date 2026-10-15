# monitoring

Provisions platform observability:

- **SLOs** for fraud scoring (latency, availability) and document
  processing (latency, error rate).
- **Alert policies** that notify on SLO burn-rate, Vertex AI drift
  detection, and BigQuery slot saturation.
- **Notification channels** (email + Pub/Sub) for fanout into on-call
  tooling.
- **Pre-built dashboards** in `dashboards/` — `vertex_ai.json`,
  `fraud.json`, `overview.json` — that visualize latency, decision
  distributions, and model health.
