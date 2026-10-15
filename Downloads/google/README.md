# FinGuard AI — AI-Powered Financial Intelligence Platform

> Real-time fraud detection, risk intelligence, document understanding, and personalized financial guidance — built on Google Cloud AI.

[![Status](https://img.shields.io/badge/status-scaffold-yellow)]()
[![License](https://img.shields.io/badge/license-Apache_2.0-blue)]()
[![Live data](https://img.shields.io/badge/live_data-verified-green)]()
[![Go tests](https://img.shields.io/badge/go_tests-18%2F18-brightgreen)]()
[![Python tests](https://img.shields.io/badge/python_tests-47%2F47-brightgreen)]()

---

## Overview

FinGuard AI unifies **fraud detection**, **risk assessment**, **document intelligence**, and **personalized financial assistance** into a single intelligent layer for banks, NBFCs, insurers, and fintechs.

The platform combines real-time streaming analytics (Pub/Sub + Dataflow + BigQuery), predictive ML (Vertex AI), generative reasoning (Gemini), and document understanding (Document AI) to reduce fraud losses, accelerate compliance, and improve customer trust.

### Capabilities

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

## Repository Layout

```
.
├── frontend/         # Next.js (TypeScript) — Web + Analyst Console + BFF API routes
├── services/         # Go microservices on Cloud Run
│   ├── cmd/          # Service entry points (fraud, risk, chat, document, compliance, api-gateway, alert-worker, db-bootstrap)
│   └── pkg/          # Shared Go packages (config, alloydb, riskstore, bigquery, vertex, domain, httpapi, …)
├── ml/               # Python ML libs, training, Vertex AI Pipelines
├── dataflow/         # Apache Beam streaming & batch jobs
├── terraform/        # GCP infrastructure as code (landing zone + app)
├── deploy/           # Cloud Build, Cloud Deploy, GitHub Actions
├── docs/             # PRD, TRD, MLOps, security, API ref, ADRs, runbooks
├── scripts/          # Local dev utilities, load tests, infra bring-up
└── .github/          # Repo-level GitHub config
```

See [`ARCHITECTURE.md`](./ARCHITECTURE.md) for the high-level system diagram and
[`CHANGES.md`](./CHANGES.md) for the most recent changes (live data integration, tradeoffs, and what was done).

---

## How the System Works

### 1. Request flow (transaction scoring)

```
Client (web / mobile / partner)
  │
  ▼  HTTPS
API Gateway  (api-gateway :8085)
  │     • JWT validation
  │     • Routing
  │     • Rate limit
  ▼
fraud-service (:8080) — Go on Cloud Run
  │
  ├──► BigQuery feature lookup  (cached velocity, geo, merchant risk)
  ├──► Vertex AI predict        (autoencoder + XGBoost ensemble, or rule fallback)
  ├──► AlloyDB alert write      (Postgres in local dev)
  └──► Pub/Sub fanout           (txn-scored)
  │
  ▼
Decision: APPROVE | REVIEW | BLOCK  (P95 < 500 ms)
```

### 2. Customer dashboard (read path)

The customer dashboard (`/customer`) and analyst console (`/analyst`) read
from two new endpoints that surface the same data as the upstream systems:

| Endpoint | BFF proxy | Backend |
|---|---|---|
| `GET /v1/alerts` | `GET /api/alerts` | fraud-service → AlloyDB `alerts` table |
| `GET /v1/customers/{id}/transactions` | `GET /api/customers/{id}/transactions` | risk-service → riskstore |

The `riskstore` package has **two interchangeable clients** wired by
`riskstore.Select(cfg)`:

- `MockClient` — deterministic, in-process, returns `simulated:true`
- `BigQueryClient` — raw HTTP against the goccy BigQuery emulator at
  `BIGQUERY_EMULATOR_HOST`, returns `simulated:false, source:"riskstore:BigQueryClient"`

Likewise, the `alloydb` package has a `MockClient` and a `RealClient`
(pgx) wired by `alloydb.Select(ctx, cfg)`. The `isSafeBQString` validator
inside `bigquery.go` guards the only `customerID` parameter that is
interpolated into a query.

### 3. Chat (LLM) flow

```
User ──► /api/chat (Next.js BFF)
            │
            ▼
        chat-bff (:8082)
            │   • session_id
            │   • PII redaction (SSN, card, email)
            ▼
        gemini_service (Python :8090)
            │   • Gemini + RAG (Vertex AI Search)
            ▼
        redacted answer + action  (e.g. verify_identity)
```

### 4. Live vs simulated — how it's decided

The `simulated` field is **derived from which client served the request**:

| Env var | When set | When unset |
|---|---|---|
| `ALLOYDB_DSN` | `alloydb.RealClient` (pgx → Postgres) | `alloydb.MockClient` |
| `BIGQUERY_EMULATOR_HOST` | `riskstore.BigQueryClient` (HTTP → goccy BQ emulator) | `riskstore.MockClient` |

- `simulated=false`, `source="alloydb.Postgres"` — `ALLOYDB_DSN` set & connect OK
- `simulated=false`, `source="riskstore:BigQueryClient"` — emulator reachable
- `simulated=true`, `source="alloydb.MockClient"` / `"riskstore:MockClient"` — fallback

The frontend (`lib/auth.ts`) keeps a `MOCK_ALERTS` / `MOCK_TXNS` fallback
that activates only when the BFF is **unreachable**, not when it returns
`simulated:true`. This is the safety net for offline development.

### 5. Infrastructure (local dev)

`scripts/restart_all.sh` is idempotent and brings up:

| Component | Port | Notes |
|---|---:|---|
| Next.js (BFF) | 3000 | dev server |
| fraud-service | 8080 | Go |
| risk-service | 8081 | Go |
| chat-bff | 8082 | Go |
| document-service | 8083 | Go |
| compliance-service | 8084 | Go |
| api-gateway | 8085 | Go |
| gemini_service | 8090 | Python / FastAPI |
| document_service | 8091 | Python / FastAPI |
| compliance_llm | 8092 | Python / FastAPI |
| Postgres | 5433 | docker-compose, alerts store |
| goccy BigQuery emulator | 9050 | binary, transactions store |

With `USE_LIVE_DATA=true` (default) it also bootstraps Postgres (5 alerts)
and seeds the BQ emulator (50 transactions for `C-1001`).

---

## Quickstart (Local Development)

> Prereqs: Go 1.22+, Python 3.11+, Node 20+, Terraform 1.6+, Docker (for
> Postgres), `bigquery-emulator` v0.6.6 (installed by `scripts/setup.sh`).

```bash
# Clone
git clone https://github.com/bikash-20/AI-BUILDER-CUP-2026.git finguard
cd finguard

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

To run a single Go service:

```bash
cd services/cmd/fraud-service
GCP_PROJECT_ID=finguard-dev MOCK_MODE=true go run .
```

To run a single Python ML service:

```bash
cd ml/services/document_service
pip install -r requirements.txt
uvicorn main:app --reload --port 8091
```

To run the frontend:

```bash
cd frontend
npm install
npm run dev          # http://localhost:3000
```

To validate the Terraform:

```bash
cd terraform/envs/dev
terraform init -backend=false
terraform validate
```

---

## Manual verification of the live path

```bash
# 1. Direct probe (should now be simulated=false)
curl -s http://127.0.0.1:8080/v1/alerts | jq '.simulated, .source'
# expect: false, "alloydb.Postgres"

curl -s http://127.0.0.1:8081/v1/customers/C-1001/transactions | jq '.simulated, .source'
# expect: false, "riskstore:BigQueryClient"

# 2. Frontend smoke
curl -s http://127.0.0.1:3000/api/alerts | jq '.simulated'
# expect: false

# 3. Full E2E (section 8 runs deep assertions)
bash _test_bff.sh
```

---

## Testing

| Layer | How | Status |
|---|---|---|
| Go services | `go test ./...` from `services/` | 18 / 18 pass |
| Go live tests | `TEST_ALLOYDB_DSN` / `TEST_BIGQUERY_EMULATOR_HOST` gated; auto-skip when unset | green online & offline |
| Python libs | `pytest libs/` from `ml/` | 47 / 47 pass |
| Frontend | `npm test` from `frontend/` | 7 / 7 pass |
| Terraform | `terraform validate` per env/module | clean |
| BFF E2E | `bash _test_bff.sh` at repo root | 35×200 + 5×202 + 7×400 + 1×404 + 1×502; section 8 deep checks live data |

---

## Target Cloud Architecture

```
┌────────────────────────────────────────────────────────────────┐
│ Clients (Web, Mobile, Analyst Console, Partner APIs)           │
└───────────────┬────────────────────────────────────────────────┘
                │ HTTPS / gRPC
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

## Tradeoffs (see also [`CHANGES.md`](./CHANGES.md))

| Decision | Pro | Con |
|---|---|---|
| Polyglot (Go + Python + TS) | Right tool for each job; rich ML ecosystem in Python; low-latency hot path in Go | More toolchains; cross-service tracing & CI care needed |
| goccy BigQuery emulator (raw HTTP) | No SDK dependency, no auth, works identically for local + real BQ (with auth header added) | Have to write a small BQ REST client; `customerID` is interpolated → must validate with `isSafeBQString` |
| `pgx` for local Postgres (instead of full AlloyDB driver) | Wire-compatible with AlloyDB; lightweight; existing pg mock tests still apply | Connection-pool tuning is local-dev tuned, not prod |
| Mock fallback in service layer (not just BFF) | Service still boots if Postgres/BQ emulator is down; `simulated` flag is honest | Operators have to read the log line to know why data is simulated |
| Live data tests gated on env vars | CI stays fast & hermetic | Risk of regression on the live path if you only run offline tests |
| Frontend `MOCK_*` fallback only on BFF unreachable | Avoids masking real outages; users see real `simulated:true` rows | A real outage shows a static fallback; we considered a banner but it adds noise |

---

## Documentation Index

- 🏗 [`ARCHITECTURE.md`](./ARCHITECTURE.md) — Layered architecture, service responsibilities, MLOps lifecycle
- 🛠 [`RUNBOOK_LOCAL.md`](./RUNBOOK_LOCAL.md) — Local dev status snapshot & how-to
- 🆕 [`CHANGES.md`](./CHANGES.md) — Recent changes, tradeoffs, what was done, critical parts
- 📘 [`docs/prd.md`](./docs/prd.md) — Product Requirements Document
- 🛠 [`docs/trd.md`](./docs/trd.md) — Technical Requirements Document
- 🔄 [`docs/flows.md`](./docs/flows.md) — Application & sequence flows
- 🤖 [`docs/mlops.md`](./docs/mlops.md) — MLOps design
- 🔒 [`docs/security.md`](./docs/security.md) — Security & compliance
- 📡 [`docs/api-reference.md`](./docs/api-reference.md) — REST API contracts
- 📚 [`docs/adr/`](./docs/adr/) — Architecture Decision Records
- 🛟 [`docs/runbooks/`](./docs/runbooks/) — Incident response

---

## Contributing

PRs target `main`, require review, and run Cloud Build on every push. See
[`.github/ISSUE_TEMPLATE`](./.github/ISSUE_TEMPLATE) for bug & feature
templates.

---

## License

Apache-2.0. See [`LICENSE`](./LICENSE).
