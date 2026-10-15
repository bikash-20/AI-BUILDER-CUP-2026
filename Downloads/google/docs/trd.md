# FinGuard AI — Technical Requirements Document (TRD)

## 1. Architecture Overview

```
┌────────────────────────────────────────────────────────────────┐
│ CLIENT LAYER                                                     │
│ Web App | Mobile App | Analyst Console | Partner APIs            │
└───────────────┬────────────────────────────────────────────────┘
                │ HTTPS / gRPC
┌───────────────▼────────────────────────────────────────────────┐
│ API GATEWAY (Cloud Endpoints / Apigee)                            │
│ Auth (IAP + OAuth2) | Rate Limit | WAF                            │
└───────────────┬────────────────────────────────────────────────┘
                │
┌───────────────▼────────────────────────────────────────────────┐
│ APPLICATION LAYER (Cloud Run)                                     │
│ ┌──────────┬──────────┬───────────┬────────────┬───────────┐     │
│ │ Fraud    │ Document │ Chat      │ Risk       │ Compliance│     │
│ │ Service  │ Service  │ Service   │ Service    │ Service   │     │
│ └──────────┴──────────┴───────────┴────────────┴───────────┘     │
└───────┬──────────────┬──────────────┬──────────────┬─────────────┘
        │              │              │              │
┌───────▼──────┐ ┌─────▼──────┐ ┌─────▼─────┐ ┌──────▼─────────┐
│ Pub/Sub      │ │ Vertex AI  │ │ Gemini    │ │ Document AI    │
│ (streaming)  │ │ (ML)       │ │ (LLM)     │ │ (OCR + NER)    │
└───────┬──────┘ └─────┬──────┘ └─────┬─────┘ └──────┬─────────┘
        │              │              │              │
┌───────▼──────────────▼──────────────▼──────────────▼────────────┐
│ DATA LAYER                                                         │
│ BigQuery (analytics) | AlloyDB (OLTP) | Cloud Storage (docs)       │
│ Vertex AI Search (vector) | Memorystore (cache) | Pub/Sub DLQ      │
└────────────────────────────────────────────────────────────────────┘
                │
┌───────────────▼────────────────────────────────────────────────┐
│ OBSERVABILITY & SECURITY                                            │
│ Cloud Logging | Monitoring | Trace | SCC | DLP | KMS | IAM         │
└────────────────────────────────────────────────────────────────────┘
```

## 2. Google Cloud Services Mapping

| Layer | Service | Purpose |
|---|---|---|
| Ingest | Pub/Sub | Transaction & event streaming |
| Ingest | Cloud Storage | Document & batch ingestion |
| Processing | Dataflow | Stream ETL & enrichment |
| Analytics | BigQuery | Feature store, risk analytics |
| OLTP | AlloyDB / Cloud SQL | App data, cases, users |
| ML | Vertex AI | Training, prediction, pipelines |
| LLM | Gemini on Vertex AI | Reasoning, chat, summarization |
| Docs | Document AI | OCR, NER, form parsing |
| Search | Vertex AI Search | Policy/knowledge retrieval |
| Apps | Cloud Run | Stateless microservices |
| Auth | IAP + Identity Platform | User auth |
| Secrets | Secret Manager | API keys |
| Security | VPC-SC, CMEK, DLP, SCC | Compliance |
| Observability | Cloud Logging/Monitoring/Trace | SRE |

## 3. Functional Requirements

### FR-1: Transaction Ingestion
- Accept events from core banking via Pub/Sub
- Schema: `{txn_id, customer_id, amount, currency, merchant, geo, device, timestamp}`
- Guarantee at-least-once delivery
- Dead-letter queue for failures

### FR-2: Real-Time Scoring
- Enrich with BigQuery features (velocity, geo-velocity, merchant risk)
- Call Vertex AI endpoint for anomaly score
- Return decision: `APPROVE | REVIEW | BLOCK` in < 500 ms (P95)

### FR-3: Document Processing
- Accept PDF/JPG/PNG via signed URL
- Document AI extractor → structured JSON
- Validate against KYC/AML rules
- Store in Cloud Storage + AlloyDB

### FR-4: Conversational Assistant
- Gemini-powered chat with RAG over Vertex AI Search
- Context: customer profile, transactions, policies
- Guardrails: PII masking, toxicity filter, citation enforcement

### FR-5: Risk Scoring
- Batch (nightly) + real-time (per event)
- Composite score: fraud + credit + behavioral + compliance
- Exposed via REST API

### FR-6: Compliance Automation
- Auto-generate SAR/CTR drafts
- Rule engine + LLM summarization
- Immutable audit log in BigQuery (append-only)

## 4. Non-Functional Requirements

| Category | Requirement |
|---|---|
| Latency | P95 < 500 ms (scoring), < 2 s (chat) |
| Throughput | 50,000 TPS peak |
| Availability | 99.95 % |
| Durability | 99.999999 % (GCS) |
| Scalability | Horizontal via Cloud Run autoscaling |
| Security | Zero-trust, VPC-SC, CMEK, IAM least privilege |
| Compliance | PCI-DSS, SOC2, GDPR, RBI/SEBI (India), HIPAA (insurer) |
| Observability | Logs, metrics, traces, SLOs |
| DR | RPO < 5 min, RTO < 30 min (multi-region) |

## 5. Data Model (Key Tables in BigQuery)

See [`terraform/modules/data/schemas/`](../terraform/modules/data/schemas) for canonical DDL.

## 6. API Contracts (REST)

See [`docs/api-reference.md`](./api-reference.md) for the full reference; OpenAPI spec at [`terraform/modules/apigateway/openapi.yaml`](../terraform/modules/apigateway/openapi.yaml).

## 7. ML / AI Design

See [`docs/mlops.md`](./mlops.md) for the full design — feature store, training, evaluation, registration, deployment, monitoring, drift, CT pipeline.

## 8. Security & Compliance

See [`docs/security.md`](./security.md).

## 9. Deployment & CI/CD

See [`deploy/`](../deploy) for Cloud Build and Cloud Deploy configs.

## 10. Observability & SLOs

| SLI | SLO |
|---|---|
| Scoring availability | 99.95 % |
| P95 scoring latency | < 500 ms |
| Chat success rate | > 99 % |
| Doc extraction accuracy | > 95 % |
| Error budget | 0.05 % / month |
