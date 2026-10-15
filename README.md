# FinGuard AI — AI-Powered Financial Intelligence Platform

> Real-time fraud detection, risk intelligence, document understanding, and personalized financial guidance — built on Google Cloud AI.

[![Status](https://img.shields.io/badge/status-scaffold-yellow)]()
[![License](https://img.shields.io/badge/license-Apache_2.0-blue)]()
[![Live data](https://img.shields.io/badge/live_data-verified-green)]()
[![Go tests](https://img.shields.io/badge/go_tests-18%2F18-brightgreen)]()
[![Python tests](https://img.shields.io/badge/python_tests-47%2F47-brightgreen)]()

---

## Where is the code?

The entire project lives in **[`Downloads/google/`](./Downloads/google/)**.

This top-level `README.md` is the landing page. All sources, tests, scripts,
documentation, and infrastructure are in the inner folder.

```text
.
├── README.md                                 ← you are here
└── Downloads/
    └── google/                               ← the actual project
        ├── README.md                         ← project landing page
        ├── ARCHITECTURE.md                   ← high-level architecture
        ├── ARCHITECTURE_DEEP_DIVE.md         ← full teaching walk-through
        ├── CHANGES.md                        ← recent changes & tradeoffs
        ├── RUNBOOK_LOCAL.md                  ← local dev runbook
        ├── Makefile
        ├── docker-compose.yml
        ├── _test_bff.sh                      ← end-to-end BFF test
        │
        ├── frontend/                         ← Next.js (TypeScript)
        ├── services/                         ← Go 1.22+ microservices
        ├── ml/                               ← Python ML ecosystem
        ├── dataflow/                         ← Apache Beam jobs
        ├── terraform/                        ← Infrastructure as code
        ├── deploy/                           ← Cloud Build / Cloud Deploy / GitHub Actions
        ├── docs/                             ← PRD, TRD, MLOps, Security, ADRs, Runbooks
        └── scripts/                          ← Local dev utilities
```

---

## What is FinGuard AI?

**FinGuard AI** is a real-time financial intelligence platform for banks,
NBFCs, insurers, and fintechs. It unifies four products in one:

1. **Real-time fraud detection** on transactions (sub-500 ms P95).
2. **Composite risk scoring** per customer (0–1000).
3. **Document understanding** for KYC forms, claims, and invoices (Document AI).
4. **Conversational assistant** with PII redaction (Gemini + RAG).

Plus it auto-drafts compliance filings (SAR / CTR) for the compliance team.

Everything runs on Google Cloud. The hot path is Go on Cloud Run; the
ML / LLM path is Python on Cloud Run; the frontend is Next.js; the data
lives in BigQuery, AlloyDB, and Cloud Storage.

---

## Highlights

| Capability | Implementation |
|---|---|
| Real-time fraud | Vertex AI autoencoder + XGBoost ensemble, sub-500 ms scoring on Cloud Run |
| Document AI | Prebuilt + custom Document AI extractors for KYC, claims, invoices |
| Conversational assistant | Gemini on Vertex AI with RAG over Vertex AI Search |
| Risk scoring | Composite 0–1000 score (fraud + credit + behavioral + compliance) |
| Compliance automation | LLM-drafted SAR/CTR, immutable audit trail in BigQuery |
| MLOps | Vertex AI Pipelines (Kubeflow), Model Registry, drift monitoring |
| Live data path | Real Postgres (alerts) + goccy BigQuery emulator (transactions) behind the same APIs |

---

## Quickstart

```bash
# Clone
git clone https://github.com/bikash-20/AI-BUILDER-CUP-2026.git finguard
cd finguard/Downloads/google

# First-time setup (installs bigquery-emulator, checks docker)
./scripts/setup.sh

# Bring everything up (Postgres + BQ emulator + 10 services + frontend)
./scripts/restart_all.sh

# Verify
./scripts/status.sh
./_test_bff.sh     # E2E; section 8 runs deep live-data assertions
```

To run in **mock-only** mode (no Postgres, no BQ emulator):

```bash
USE_LIVE_DATA=false ./scripts/restart_all.sh
```

---

## Documentation

Start here, in this order:

1. 📘 **[`Downloads/google/README.md`](./Downloads/google/README.md)** — project landing page
2. 🏗 **[`Downloads/google/ARCHITECTURE.md`](./Downloads/google/ARCHITECTURE.md)** — layered architecture, service responsibilities, MLOps lifecycle
3. 🎓 **[`Downloads/google/ARCHITECTURE_DEEP_DIVE.md`](./Downloads/google/ARCHITECTURE_DEEP_DIVE.md)** — full teaching walk-through (1,000+ lines, 19 sections)
4. 🆕 **[`Downloads/google/CHANGES.md`](./Downloads/google/CHANGES.md)** — recent changes, tradeoffs, what was done
5. 🛠 **[`Downloads/google/RUNBOOK_LOCAL.md`](./Downloads/google/RUNBOOK_LOCAL.md)** — local dev runbook
6. 📚 **[`Downloads/google/docs/`](./Downloads/google/docs/)** — PRD, TRD, MLOps, Security, API ref, ADRs, Runbooks

---

## Target Cloud Architecture

```
┌────────────────────────────────────────────────────────────────┐
│ Clients (Web, Mobile, Analyst Console, Partner APIs)           │
└───────────────┬────────────────────────────────────────────────┘
                │ HTTPS
┌───────────────▼────────────────────────────────────────────────┐
│ API Gateway (Apigee / Cloud Endpoints, IAP)                     │
└───────────────┬────────────────────────────────────────────────┘
                │
┌───────────────▼────────────────────────────────────────────────┐
│ Cloud Run (Go + Python): fraud • risk • document • chat • compliance │
└─────┬───────────┬──────────────────┬──────────────────────┬─────┘
      │           │                  │                      │
      ▼           ▼                  ▼                      ▼
   Pub/Sub    Vertex AI         Gemini (LLM)        Document AI
   (stream)   (ML)              + RAG                (OCR/NER)
      │           │                  │                      │
      ▼           ▼                  ▼                      ▼
┌────────────────────────────────────────────────────────────────┐
│ Data: BigQuery (analytics) • AlloyDB (OLTP) • GCS • Vertex AI Search │
└────────────────────────────────────────────────────────────────┘
                │
┌───────────────▼────────────────────────────────────────────────┐
│ Observability & Security: Logging • Monitoring • Trace • SCC • DLP │ KMS • IAM │
└────────────────────────────────────────────────────────────────┘
```

Primary region: **asia-south1 (Mumbai)** — DR: **asia-southeast1**.

---

## Key Metrics & Targets

| Metric | Target |
|---|---|
| Fraud detection recall | > 95 % |
| False positive rate | < 5 % |
| P95 scoring latency | < 500 ms |
| P95 chat latency | < 2 s |
| Doc extraction time | < 30 s / doc |
| Availability | 99.95 % |
| RPO / RTO | < 5 min / < 30 min |

---

## Testing

| Layer | How | Status |
|---|---|---|
| Go services | `go test ./...` from `Downloads/google/services/` | 18 / 18 pass |
| Go live tests | `TEST_ALLOYDB_DSN` / `TEST_BIGQUERY_EMULATOR_HOST` gated; auto-skip when unset | green online & offline |
| Python libs | `pytest libs/` from `Downloads/google/ml/` | 47 / 47 pass |
| Frontend | `npm test` from `Downloads/google/frontend/` | 7 / 7 pass |
| Terraform | `terraform validate` per env/module | clean |
| BFF E2E | `bash _test_bff.sh` at `Downloads/google/` | 35×200 + 5×202 + 7×400 + 1×404 + 1×502; section 8 deep checks live data |

---

## License

Apache-2.0. See [`Downloads/google/LICENSE`](./Downloads/google/LICENSE).
