# Runbook — Fraud Spike

> Triggered when alert volume exceeds 3× the rolling 7-day baseline, or P95 scoring latency > 1 s for 5 minutes.

## 1. Detect

- Looker dashboard: "Fraud — Live KPIs"
- Alert: `fraud_alert_rate > 3x_baseline` (Cloud Monitoring)
- Alert: `scoring_latency_p95 > 1000ms` for 5 min

## 2. Triage (first 5 min)

1. **Identify scope** — which merchants, geos, customer segments are spiking?
2. **Check feature skew** — Vertex AI Model Monitoring dashboard
3. **Check upstream** — are we seeing a real attack, or has data distribution changed?
4. **Page on-call ML engineer** if not auto-paged

## 3. Mitigate (next 15 min)

| Symptom | Action |
|---|---|
| Latency spike from Vertex AI | Switch `fraud-service` to **rule-based fallback** (`MOCK_VERTEX=true` for the deployment) |
| False positives spike | Raise `REVIEW_THRESHOLD` from 0.7 to 0.85 via feature flag |
| Confirmed novel attack | Add rule in [`services/cmd/fraud-service/decision.go`](../../services/cmd/fraud-service/decision.go) and redeploy via Cloud Deploy canary |
| Data quality issue | Pause fraud-service ingestion; engage Dataflow on-call |

## 4. Communicate

- Slack `#fraud-incidents` with:
  - Time of spike
  - Affected segments
  - Mitigation in place
  - ETA for full recovery
- Status page: "Investigating elevated fraud alerts"

## 5. Post-incident

- Within 24 h: postmortem doc, action items
- Within 7 days: model retraining if drift confirmed
- Update runbook with new findings
