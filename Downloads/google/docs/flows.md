# FinGuard AI — Application & Sequence Flows

## 1. High-Level End-to-End Flow

```
[Customer / Analyst / System]
        │
        ▼
[API Gateway + Auth]
        │
        ▼
[Cloud Run Microservices]
        │
        ├──► Fraud Service ──► Pub/Sub ──► Dataflow ──► BigQuery
        │                         │
        │                         ▼
        │                    Vertex AI (score)
        │                         │
        │                         ▼
        │                    Decision Engine
        │
        ├──► Document Service ──► Document AI ──► AlloyDB
        │
        ├──► Chat Service ──► Gemini + Vertex AI Search
        │
        ├──► Risk Service ──► BigQuery + Vertex AI
        │
        └──► Compliance Service ──► Rules + LLM + Audit Log
```

## 2. Flow 1 — Real-Time Fraud Detection

```
1. Customer swipes card / initiates txn
2. Core banking publishes event → Pub/Sub topic: txn-events
3. Dataflow job consumes, enriches with BigQuery features
4. Enriched event → Cloud Run Fraud Service
5. Fraud Service calls Vertex AI endpoint → fraud_probability
6. Decision engine applies thresholds:
      < 0.3   → APPROVE
      0.3–0.7 → REVIEW
      > 0.7   → BLOCK
7. If REVIEW/BLOCK:
      - Create alert in AlloyDB
      - Notify analyst console (WebSocket)
      - Optionally SMS/push customer
8. Response returned to core banking in < 500 ms
9. All events logged to BigQuery for retraining
```

## 3. Flow 2 — Document Intelligence (KYC / Loan / Claim)

```
1. Customer uploads doc via app → signed URL → Cloud Storage
2. Cloud Storage event → Pub/Sub → Document Service
3. Document Service calls Document AI processor
4. Extracted JSON → validation rules (KYC/AML)
5. If confidence < 0.9 → human review queue
6. If valid → store in AlloyDB + BigQuery
7. Notify customer via chat/push
8. Audit log written
```

## 4. Flow 3 — Conversational Assistant

```
1. User sends message → Chat Service
2. DLP redacts PII
3. Chat Service builds context:
      - Customer profile (AlloyDB)
      - Recent txns (BigQuery)
      - Policies (Vertex AI Search)
4. Gemini generates response with citations
5. Guardrails: toxicity, hallucination, PII leak check
6. If action intent → function call (e.g., raise_dispute)
7. Response returned with citations + actions
8. Session logged (redacted) for QA
```

## 5. Flow 4 — Risk Scoring (Batch + Real-Time)

```
Real-time:
  Event → Fraud Service → Vertex AI → risk_score → AlloyDB

Batch (nightly):
  Cloud Scheduler → Dataflow → BigQuery → Vertex AI batch predict
  → risk_scores table → Looker dashboards
```

## 6. Flow 5 — Compliance Automation

```
1. Fraud alert OR periodic trigger
2. Compliance Service pulls case data from AlloyDB + BigQuery
3. Rules engine evaluates AML/KYC criteria
4. Gemini summarizes case → draft SAR/CTR
5. Compliance officer reviews + approves
6. Immutable audit log written to BigQuery
7. Report filed via regulator API
```

## 7. Sequence Diagram (Fraud Path)

```
Customer   CoreBank   PubSub   Dataflow   FraudSvc   VertexAI   AlloyDB   Analyst
   │          │         │        │          │          │          │         │
   │ swipe    │         │        │          │          │          │         │
   ├─────────►│         │        │          │          │          │         │
   │          │ publish │        │          │          │          │         │
   │          ├────────►│        │          │          │          │         │
   │          │         │ consume│          │          │          │         │
   │          │         ├───────►│          │          │          │         │
   │          │         │        │ enrich   │          │          │         │
   │          │         │        ├─────────►│          │          │         │
   │          │         │        │          │ predict  │          │         │
   │          │         │        │          ├─────────►│          │         │
   │          │         │        │          │◄─────────┤          │         │
   │          │         │        │          │ decision │          │         │
   │          │         │        │          ├─────────────────────►│        │
   │          │         │        │          │          │          │ alert  │
   │          │         │        │          │          │          ├───────►│
   │          │◄────────────────────────────┤          │          │         │
   │◄─────────┤         │        │          │          │          │         │
```

## 8. Error Handling & Fallbacks

| Failure | Fallback |
|---|---|
| Vertex AI timeout | Rule-based score |
| Document AI low confidence | Human review queue |
| Gemini unavailable | Cached FAQ response |
| Pub/Sub backlog | Autoscale Dataflow |
| BigQuery quota | Serve from Memorystore cache |

## 9. Deployment Topology

- Region: `asia-south1` (primary), `asia-southeast1` (DR)
- Multi-zone: Cloud Run + AlloyDB HA
- Global LB: Cloud Load Balancing + Cloud CDN
- DR: cross-region replication, RPO < 5 min
