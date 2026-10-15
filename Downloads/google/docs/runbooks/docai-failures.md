# Runbook — Document AI Failures

> Triggered when Document AI extraction success rate drops below 90 %, or queue depth exceeds 1,000.

## 1. Detect

- Cloud Monitoring: `docai_success_rate < 0.9` for 10 min
- Cloud Monitoring: `docai_queue_depth > 1000`
- Looker: "Doc AI — Health"

## 2. Triage

1. **Sample failures** — get 10 failed docs from DLQ (`gs://finguard-docai-dlq/`)
2. **Categorize**:
   - Blurry / unscannable
   - Unsupported format
   - Processor error
   - Timeout

3. **Check processor** — Document AI processor status, region availability

## 3. Mitigate

| Failure | Action |
|---|---|
| Blurry / unscannable | Push to **human-in-the-loop** review queue; notify customer |
| Unsupported format | Add format to allowed list; document customer comms |
| Processor error | Failover to alternative processor (form → identity) |
| Timeout | Increase client timeout; reduce batch size |
| Quota | Request quota increase; rate-limit intake |

## 4. Human-in-the-loop

Failed docs land in **review queue** (AlloyDB table `documents.human_review`).

- SLA: 4 h during business hours, 24 h off-hours
- Escalation: after 24 h, ping ops lead

## 5. Post-incident

- Weekly review of failure patterns
- Improve preprocessing (image enhancement, de-skew) in `ml/libs/finguard_docai/`
- Adjust confidence threshold (currently 0.9) if too aggressive
