# FinGuard AI — Architecture

This document explains the layered architecture, key decisions, and runtime topology of the FinGuard AI platform.

## Layered View

```
┌────────────────────────────────────────────────────────────────────┐
│  CLIENT LAYER                                                       │
│  - Web App (Next.js)                                                │
│  - Mobile App (later phase)                                         │
│  - Analyst Console (role-based)                                     │
│  - Partner APIs                                                      │
└─────────────────────────────┬──────────────────────────────────────┘
                              │ HTTPS / gRPC
┌─────────────────────────────▼──────────────────────────────────────┐
│  EDGE / API GATEWAY                                                  │
│  - Cloud Endpoints or Apigee                                         │
│  - Identity-Aware Proxy (IAP) + OAuth2                               │
│  - Cloud Armor (WAF), rate limiting                                  │
└─────────────────────────────┬──────────────────────────────────────┘
                              │
┌─────────────────────────────▼──────────────────────────────────────┐
│  APPLICATION LAYER (Cloud Run)                                      │
│  ┌──────────────┬──────────────┬──────────────┬──────────────┐    │
│  │ fraud-       │ risk-        │ document-    │ chat-bff     │    │
│  │ service (Go) │ service (Go) │ service (Go) │ (Go)         │    │
│  ├──────────────┼──────────────┼──────────────┼──────────────┤    │
│  │ compliance-  │ alert-       │ document_    │ gemini_      │    │
│  │ service (Go) │ worker (Go)  │ service      │ service      │    │
│  │              │              │ (Python)     │ (Python)     │    │
│  └──────────────┴──────────────┴──────────────┴──────────────┘    │
└─────┬──────────┬───────────┬───────────┬────────────┬──────────────┘
      │          │           │           │            │
      ▼          ▼           ▼           ▼            ▼
┌──────────┐ ┌────────┐ ┌────────┐ ┌──────────┐ ┌──────────┐
│ Pub/Sub  │ │Vertex AI│ │ Gemini │ │Document AI│ │ Vertex   │
│(streaming)│ │ (ML)    │ │ (LLM)  │ │ (OCR/NER)│ │ AI Search│
└────┬─────┘ └───┬────┘ └────┬───┘ └─────┬────┘ └────┬─────┘
     │           │           │            │           │
     ▼           ▼           ▼            ▼           ▼
┌────────────────────────────────────────────────────────────────────┐
│  DATA LAYER                                                         │
│  - BigQuery (analytics, feature store, audit log)                   │
│  - AlloyDB / Cloud SQL (OLTP: alerts, cases, users, sessions)       │
│  - Cloud Storage (documents, model artifacts)                       │
│  - Vertex AI Feature Store (online + offline)                       │
│  - Memorystore (Redis cache, fallback)                              │
│  - Pub/Sub DLQ                                                       │
└────────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────▼──────────────────────────────────────┐
│  CROSS-CUTTING: Observability, Security, Compliance                  │
│  - Cloud Logging + Monitoring + Trace                                │
│  - Error Reporting, SLOs                                             │
│  - Security Command Center, VPC-SC, CMEK, DLP                       │
│  - Cloud Audit Logs → BigQuery append-only                           │
└────────────────────────────────────────────────────────────────────┘
```

## Service Responsibilities

| Service | Language | Primary Endpoint(s) | Notes |
|---|---|---|---|
| `api-gateway` | Go | `*` (proxy) | JWT validation, rate limit, routing |
| `fraud-service` | Go | `POST /v1/score`, `POST /v1/transactions` | Sub-500 ms with Vertex AI + rule fallback |
| `risk-service` | Go | `GET /v1/risk/{customer_id}`, `POST /v1/risk/recompute` | Composite 0–1000 score |
| `document-service` | Go | `POST /v1/documents/upload`, `GET /v1/documents/{id}` | BFF for Python `document_service` |
| `chat-bff` | Go | `POST /v1/chat` | Session mgmt + DLP redaction → Python `gemini_service` |
| `compliance-service` | Go | `POST /v1/compliance/sar/draft` | Audit log + SAR/CTR routing |
| `alert-worker` | Go | (background) | Pub/Sub subscriber → WebSocket fanout |
| `document_service` | Python (FastAPI) | `POST /extract`, `POST /validate` | Document AI extractors |
| `gemini_service` | Python (FastAPI) | `POST /chat`, `POST /complete` | Gemini + RAG |
| `risk_batch` | Python (Job) | (job) | Nightly batch risk scoring |
| `compliance_llm` | Python (FastAPI) | `POST /draft_sar`, `POST /draft_ctr` | LLM-drafted filings |

## Data Flow — Real-Time Fraud

```
Core Banking
    │
    ▼
Pub/Sub: txn-events
    │
    ▼
Dataflow (txn_enrichment)
    │  ◄── BigQuery features (velocity, geo-velocity, merchant risk)
    │
    ▼
Pub/Sub: txn-scoring
    │
    ▼
fraud-service (Cloud Run, Go)
    │  ◄── Vertex AI endpoint (autoencoder + XGBoost ensemble)
    ▼
Decision (APPROVE | REVIEW | BLOCK)
    │        │
    │        ▼
    │   alerts (AlloyDB) + analyst console (WebSocket)
    ▼
Response to Core Banking (< 500 ms P95)
```

## Cross-Cutting Concerns

- **Auth:** IAP + Identity Platform. OAuth2 access tokens; service-to-service via Workload Identity.
- **Secrets:** Secret Manager, never in env or code.
- **Encryption:** CMEK via KMS for BigQuery, GCS, AlloyDB, Pub/Sub, Vertex AI.
- **Network:** Hub-and-spoke VPC, Private Service Connect, Cloud NAT, no public IPs in prod.
- **Audit:** Cloud Audit Logs streamed to BigQuery (append-only table), 7-year retention.
- **Observability:** Structured JSON logs, OpenTelemetry traces, SLOs (availability, latency, error rate, model drift).

## MLOps Lifecycle

```
Pub/Sub + GCS ──► Dataflow ──► BigQuery ──► Feature Store
                                              │
                                              ▼
                              Vertex AI Pipelines (Kubeflow)
                              ┌─ Validate ─► Features ─► Train ─► Evaluate ─► Register ─► Deploy
                                              │
                                              ▼
                              Vertex AI Endpoint (online) + Batch Prediction
                                              │
                                              ▼
                              Model Monitoring (drift, skew, latency)
                                              │
                                              ▼
                              Auto-trigger retrain or alert
```

See [`docs/mlops.md`](./docs/mlops.md) for full design.

## Region & DR

- **Primary:** `asia-south1` (Mumbai) — data residency for Indian customers.
- **DR:** `asia-southeast1` (Singapore) — cross-region replication.
- **RPO < 5 min, RTO < 30 min** — BigQuery snapshots, AlloyDB HA, GCS dual-region, Vertex AI multi-region endpoints.

## Trust Boundary

```
┌────────────────────────────────────────────────────────────┐
│ VPC-SC Perimeter                                            │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐     │
│  │ BigQuery     │   │ Vertex AI    │   │ Document AI  │     │
│  │ + KMS        │   │ + KMS        │   │ + KMS        │     │
│  └──────────────┘   └──────────────┘   └──────────────┘     │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐     │
│  │ Cloud Storage│   │ AlloyDB      │   │ Cloud Run    │     │
│  │ + KMS        │   │ + KMS        │   │ (services)   │     │
│  └──────────────┘   └──────────────┘   └──────────────┘     │
└────────────────────────────────────────────────────────────┘
```
