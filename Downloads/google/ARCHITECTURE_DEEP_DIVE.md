# FinGuard AI — A Deep-Dive Walk-Through

> A teaching guide. Read this top-to-bottom and you will understand how
> the entire platform fits together, why each piece exists, what data
> flows where, and the tradeoffs the team made.
>
> Audience: a developer or AI agent who has never seen the codebase.
> Time to read: 30–45 minutes.

---

## Table of contents

1. [The product in one paragraph](#1-the-product-in-one-paragraph)
2. [The five things this platform must do](#2-the-five-things-this-platform-must-do)
3. [The polyglot stack and why](#3-the-polyglot-stack-and-why)
4. [High-level architecture](#4-high-level-architecture)
5. [Repository layout, file by file](#5-repository-layout-file-by-file)
6. [The request lifecycle: a real example](#6-the-request-lifecycle-a-real-example)
7. [The Go services in detail](#7-the-go-services-in-detail)
8. [The Python services in detail](#8-the-python-services-in-detail)
9. [The data layer](#9-the-data-layer)
10. [The BFF (Next.js) layer](#10-the-bff-nextjs-layer)
11. [Configuration and the `MOCK_MODE` contract](#11-configuration-and-the-mock_mode-contract)
12. [The "live data" plan, in plain English](#12-the-live-data-plan-in-plain-english)
13. [MLOps: how models are trained, deployed, and monitored](#13-mlops-how-models-are-trained-deployed-and-monitored)
14. [Security and compliance](#14-security-and-compliance)
15. [Infrastructure as code](#15-infrastructure-as-code)
16. [CI/CD](#16-cicd)
17. [Testing strategy](#17-testing-strategy)
18. [Tradeoffs and "what we'd do differently next time"](#18-tradeoffs-and-what-wed-do-differently-next-time)
19. [Glossary](#19-glossary)

---

## 1. The product in one paragraph

**FinGuard AI** is a real-time financial intelligence platform for banks,
NBFCs, insurers, and fintechs. It does four things in one product:

1. **Catches fraud** on transactions in under 500 ms.
2. **Scores risk** for a customer as a single 0–1000 number.
3. **Reads and understands documents** (KYC forms, claims, invoices) via Document AI.
4. **Answers customer questions** in a chat window using Gemini with retrieval-augmented generation (RAG).

Plus it auto-drafts compliance filings (SAR / CTR) for the compliance team.

Everything runs on Google Cloud. The hot path is Go on Cloud Run; the
ML / LLM path is Python on Cloud Run; the frontend is Next.js; the data
lives in BigQuery, AlloyDB, and Cloud Storage.

---

## 2. The five things this platform must do

These come from `docs/prd.md` and `docs/trd.md`. The numbers map to
"FR-N" in the requirements:

| # | Requirement | How we deliver it |
|---|---|---|
| FR-1 | Ingest transactions in real time | Pub/Sub topic `txn-events` → Dataflow `txn_enrichment` |
| FR-2 | Score each transaction in < 500 ms P95 | `fraud-service` (Go) → Vertex AI autoencoder + XGBoost |
| FR-3 | Extract structured data from documents | Python `document_service` → Document AI (prebuilt + custom) |
| FR-4 | Conversational assistant with PII redaction | `chat-bff` (Go) → Python `gemini_service` → Gemini + Vertex AI Search RAG |
| FR-5 | Composite risk score per customer | `risk-service` (Go) — fraud + credit + behavioral + compliance, 0–1000 |
| FR-6 | Compliance: draft SAR / CTR filings | `compliance-service` (Go) + Python `compliance_llm` |
| FR-7 | Live data for analyst + customer dashboards | `GET /v1/alerts`, `GET /v1/customers/{id}/transactions` (Postgres + BigQuery) |

Non-functional targets: 99.95% availability, P95 < 500 ms for scoring,
RPO < 5 min, RTO < 30 min. DR is cross-region.

---

## 3. The polyglot stack and why

This is a deliberate decision, captured in `docs/adr/0001-polyglot-stack.md`.
The short version:

| Layer | Language | Why this language |
|---|---|---|
| Frontend | TypeScript + Next.js | Type safety, server-side rendering, large ecosystem |
| Cloud Run services (hot path) | **Go 1.22+** | Sub-100 ms cold start, goroutine model for fan-out, single static binary |
| ML / AI services | **Python 3.11+** | First-class Google AI SDKs, Beam, scikit-learn, Vertex AI SDK |
| Data pipelines | **Python (Beam)** | Apache Beam native, runs on Dataflow |
| IaC | **Terraform 1.6+** | GCP provider maturity, state management, monorepo-friendly |
| Frontend BFF API routes | **Next.js Route Handlers (TypeScript)** | Lives with the frontend, no separate service needed |

The cost of polyglot: more toolchains, more CI, more observability care.
The benefit: the right tool for each job. We accepted the cost.

---

## 4. High-level architecture

```
┌────────────────────────────────────────────────────────────────────┐
│  CLIENT LAYER                                                       │
│  - Web App (Next.js)                                                │
│  - Mobile App (later)                                                │
│  - Analyst Console (role-based)                                      │
│  - Partner APIs                                                      │
└─────────────────────────────┬──────────────────────────────────────┘
                              │ HTTPS
┌─────────────────────────────▼──────────────────────────────────────┐
│  EDGE / API GATEWAY                                                  │
│  - API Gateway (Go) :8085     ← we built our own (production: Apigee)│
│  - JWT validation, rate limit, routing                              │
└─────────────────────────────┬──────────────────────────────────────┘
                              │
┌─────────────────────────────▼──────────────────────────────────────┐
│  APPLICATION LAYER (Cloud Run)                                      │
│  ┌──────────────┬──────────────┬──────────────┬──────────────┐     │
│  │ fraud-       │ risk-        │ document-    │ chat-bff     │     │
│  │ service (Go) │ service (Go) │ service (Go) │ (Go)         │     │
│  ├──────────────┼──────────────┼──────────────┼──────────────┤     │
│  │ compliance-  │ alert-       │ document_    │ gemini_      │     │
│  │ service (Go) │ worker (Go)  │ service      │ service      │     │
│  │              │              │ (Python)     │ (Python)     │     │
│  └──────────────┴──────────────┴──────────────┴──────────────┘     │
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

Primary region: `asia-south1` (Mumbai). DR: `asia-southeast1`.

---

## 5. Repository layout, file by file

```
.
├── frontend/                          # Next.js 14 (App Router) — Web + Analyst Console
│   ├── src/
│   │   ├── app/
│   │   │   ├── layout.tsx            # Root layout (theme provider, role switcher)
│   │   │   ├── page.tsx              # Landing page (role-based redirect)
│   │   │   ├── (auth)/
│   │   │   │   ├── login/page.tsx    # OAuth2 login (Identity Platform)
│   │   │   │   └── callback/route.ts # OAuth2 callback handler
│   │   │   ├── customer/             # Customer dashboard
│   │   │   │   ├── page.tsx          # Risk gauge, recent transactions
│   │   │   │   ├── chat/page.tsx     # Chat with Gemini (BFF)
│   │   │   │   └── disputes/page.tsx # Open a dispute
│   │   │   ├── analyst/              # Analyst console
│   │   │   │   ├── page.tsx          # Alert queue
│   │   │   │   ├── cases/[id]/page.tsx # Case detail + timeline
│   │   │   │   ├── documents/page.tsx  # Document review
│   │   │   │   └── risk/page.tsx       # Customer risk drill-down
│   │   │   └── api/                  # ⭐ BFF (Backend-For-Frontend) routes
│   │   │       ├── score/route.ts          # POST → fraud-service /v1/score
│   │   │       ├── chat/route.ts           # POST → chat-bff → gemini_service
│   │   │       ├── alerts/route.ts         # GET  → fraud-service /v1/alerts
│   │   │       ├── customers/[id]/
│   │   │       │   └── transactions/route.ts # GET → risk-service /v1/customers/{id}/transactions
│   │   │       └── healthz/route.ts        # Aggregated health check
│   │   ├── components/
│   │   │   ├── analyst/              # AlertQueue, AlertCard, RiskGauge, ShapExplainability, CaseTimeline
│   │   │   ├── customer/             # ChatWindow, TransactionList, DisputeForm
│   │   │   ├── layout/               # Sidebar, Topbar, RoleSwitcher
│   │   │   └── ui/                   # shadcn-style primitives (Button, Card, Dialog, Badge, Table)
│   │   ├── hooks/
│   │   │   ├── useAlerts.ts          # Fetch + loading/error/simulated flags
│   │   │   ├── useChat.ts            # Chat session state
│   │   │   ├── useCustomerTransactions.ts
│   │   │   └── useRiskScore.ts
│   │   ├── lib/
│   │   │   ├── api-client.ts         # Thin fetch wrapper (jget/jpost)
│   │   │   ├── auth.ts               # OAuth helpers + MOCK_ALERTS / MOCK_TXNS safety net
│   │   │   ├── format.ts             # Currency, date, severity formatters
│   │   │   ├── types.ts              # All shared TS types (Alert, Transaction, etc.)
│   │   │   └── ws.ts                 # WebSocket helper for alert-worker
│   │   └── styles/globals.css
│   ├── tests/
│   │   ├── components.test.tsx       # Vitest snapshot tests
│   │   └── wiring.test.tsx           # Wiring / state-machine tests for TransactionList etc.
│   ├── next.config.js
│   ├── tailwind.config.ts
│   ├── vitest.config.ts
│   ├── package.json
│   └── Dockerfile
│
├── services/                          # Go 1.22+ microservices
│   ├── cmd/                           # One folder per service = one main.go
│   │   ├── fraud-service/main.go      # POST /v1/score, /v1/transactions, GET /v1/alerts
│   │   ├── risk-service/main.go       # GET /v1/risk/{id}, /v1/customers/{id}/transactions
│   │   ├── chat-bff/main.go           # POST /v1/chat (PII redaction → gemini_service)
│   │   ├── document-service/main.go   # BFF for Python document_service
│   │   ├── compliance-service/main.go # POST /v1/compliance/sar/draft
│   │   ├── api-gateway/main.go        # Reverse proxy + JWT + rate limit
│   │   ├── alert-worker/main.go       # Pub/Sub subscriber → WebSocket fanout
│   │   └── db-bootstrap/main.go       # One-shot binary: applies schema + seeds alerts
│   └── pkg/                           # Shared Go packages
│       ├── config/                    # 12-factor env loading
│       ├── httpapi/                   # HTTP server + middleware + error envelope
│       ├── logging/                   # Structured logger (zerolog-like)
│       ├── tracing/                   # OpenTelemetry setup
│       ├── auth/                      # JWT verifier
│       ├── domain/                    # Shared types (Alert, Transaction, RiskScore, …)
│       ├── alloydb/                   # ⭐ Alerts store (MockClient + pgx RealClient + Select)
│       ├── bigquery/                  # Feature client (mock + real BQ)
│       ├── riskstore/                 # ⭐ Transactions store (MockClient + BQ HTTP + Select)
│       ├── pubsub/                    # Publisher + subscriber
│       ├── vertex/                    # Vertex AI predictor (mock + real)
│       └── dlp/                       # PII redactor
│
├── ml/                                # Python ML ecosystem
│   ├── libs/                          # Reusable libraries
│   │   ├── finguard_common/           # clients, config, logging, schemas
│   │   ├── finguard_dlp/              # PII redactor (tested)
│   │   ├── finguard_docai/            # Document AI extractor + validators
│   │   ├── finguard_evaluation/       # metrics + SHAP explainability
│   │   ├── finguard_gemini/           # chat, guardrails, RAG
│   │   └── finguard_vertex/           # feature_store, predictor, registry
│   ├── pipelines/                     # Vertex AI Pipelines (Kubeflow DSL)
│   │   ├── components/                # data_validation, feature_eng, train, evaluate, register, deploy
│   │   ├── fraud_training_pipeline.py
│   │   ├── fraud_continuous_training.py
│   │   ├── feature_engineering_pipeline.py
│   │   └── docai_finetune_pipeline.py
│   ├── services/                      # FastAPI services
│   │   ├── gemini_service/            # POST /chat, /complete
│   │   ├── document_service/          # POST /extract, /validate
│   │   ├── compliance_llm/            # POST /draft_sar, /draft_ctr
│   │   └── risk_batch/                # Nightly batch scoring job
│   ├── training/                      # Training entry points
│   │   ├── fraud_xgb/                 # XGBoost fraud model
│   │   ├── fraud_autoencoder/         # Autoencoder for anomaly detection
│   │   └── docai_finetune/            # Custom Document AI processor
│   └── notebooks/                     # Exploration notebooks
│
├── dataflow/                          # Apache Beam jobs
│   ├── jobs/
│   │   ├── txn_enrichment/            # Streaming: enriches txns with BQ features
│   │   ├── risk_features_batch/       # Batch: rebuilds feature store
│   │   └── doc_intake/                # Batch: ingests documents into GCS + DocAI
│   └── templates/                     # Flex templates
│
├── terraform/                         # IaC
│   ├── bootstrap/                     # Org-level bootstrap (project factory, billing)
│   ├── modules/                       # Reusable modules
│   │   ├── organization/              # Folder structure, IAM, org policies
│   │   ├── network/                   # VPC, subnets, NAT, Private Service Connect
│   │   ├── security/                  # Cloud Armor, VPC-SC, CMEK keys
│   │   ├── data/                      # BigQuery datasets + tables, GCS buckets
│   │   ├── alloydb/                   # AlloyDB cluster + instance
│   │   ├── apigateway/                # API Gateway + OpenAPI spec
│   │   ├── cloud_run/                 # Service deployment helper
│   │   ├── vertex_ai/                 # Workbench, endpoints, model registry
│   │   ├── logging/                   # Log sinks, log-based metrics
│   │   └── monitoring/                # SLOs, dashboards, alerting
│   └── envs/                          # Per-environment root modules
│       ├── dev/
│       ├── staging/
│       └── prod/
│
├── deploy/                            # CI/CD
│   ├── cloudbuild/                    # Cloud Build triggers (Go, Python, Frontend, Terraform)
│   ├── clouddeploy/                   # Cloud Deploy delivery pipeline (progressive)
│   ├── github-actions/                # PR checks (go-ci, python-ci, frontend-ci, terraform-ci)
│   └── skaffold.yaml                  # Local k8s-style dev loop
│
├── docs/                              # Documentation
│   ├── prd.md                         # Product requirements
│   ├── trd.md                         # Technical requirements
│   ├── flows.md                       # Sequence / data flow diagrams
│   ├── mlops.md                       # MLOps lifecycle
│   ├── security.md                    # Security & compliance
│   ├── api-reference.md               # REST API contracts
│   ├── adr/                           # Architecture Decision Records
│   │   ├── 0001-polyglot-stack.md
│   │   ├── 0002-go-for-low-latency.md
│   │   ├── 0003-python-for-ml.md
│   │   └── 0004-terraform-monorepo.md
│   └── runbooks/                      # Incident response
│       ├── fraud-spike.md
│       ├── model-drift.md
│       └── docai-failures.md
│
├── scripts/                           # Local dev utilities
│   ├── restart_all.sh                 # Idempotent bring-up of everything (USE_LIVE_DATA toggle)
│   ├── stop_all.sh                    # Kills all FinGuard processes
│   ├── status.sh                      # Health check for all 12 components
│   ├── setup.sh                       # Installs bigquery-emulator, checks docker
│   ├── db_bootstrap.sh                # Applies alerts schema, seeds 5 alerts
│   ├── seed_local.sh                  # BQ seeder (uses bq CLI if available)
│   ├── seed_local_http.py             # Pure-Python BQ seeder (no bq CLI needed)
│   ├── gen_proto.sh                   # (Reserved) protobuf generation
│   ├── lint_all.sh                    # Runs every linter
│   └── load_test/                     # k6 load tests
│
├── _test_bff.sh                       # ⭐ End-to-end BFF test (35×200 + 5×202 + 7×400 + 1×404 + 1×502; section 8 = deep live-data assertions)
├── docker-compose.yml                 # Postgres:16 on :5433
├── Makefile                           # Top-level dev tasks
├── README.md                          # Project landing page
├── CHANGES.md                         # Recent changes, tradeoffs, what was done
├── RUNBOOK_LOCAL.md                   # Local dev status & how-to
├── ARCHITECTURE.md                    # Original architecture overview
├── ARCHITECTURE_DEEP_DIVE.md          # ⭐ This file
└── LICENSE                            # Apache-2.0
```

---

## 6. The request lifecycle: a real example

Let's trace **"a customer gets a fraud alert"** end to end. This will
hit most of the system.

### Step 1 — Transaction arrives

The core banking system publishes a transaction to Pub/Sub topic
`txn-events` (or, in local dev, our `pubsub.NewMockPublisher()` fakes it).

### Step 2 — Dataflow enriches it

The streaming Beam job `dataflow/jobs/txn_enrichment/pipeline.py` consumes
the event, looks up the customer's velocity / geo-velocity / merchant
risk features in BigQuery, and publishes an enriched event to Pub/Sub
`txn-scoring`.

### Step 3 — fraud-service scores it

`services/cmd/fraud-service/main.go` is a small Go HTTP server. Its
handler `handleScore`:

1. Reads the JSON body, validates fields (`json.Decoder.DisallowUnknownFields`).
2. Sets a **450 ms deadline** (the latency budget for the whole call).
3. Looks up features from `bigquery.MockClient` (or real BQ).
4. Calls `vertex.MockPredictor.Predict` (or real Vertex AI endpoint).
5. Maps the raw score to a decision: `APPROVE` / `REVIEW` / `BLOCK`
   (in `decision.go`).
6. If `REVIEW` or `BLOCK`, writes an alert to AlloyDB
   (`alloydb.RealClient.InsertAlert` or mock).
7. Returns the decision to the caller.

This is the hot path. The whole thing must finish in under 500 ms P95.
That's why it's in Go.

### Step 4 — Alert reaches the analyst console

There are two paths from the alert row to the analyst's screen:

**Path A (real-time):** `alert-worker` is a Pub/Sub subscriber on
`alerts-sub`. When a new alert row is written, it fans out to all
connected WebSocket clients. The frontend uses `lib/ws.ts` to listen.

**Path B (polling, what we have today):** The analyst console page
calls `GET /api/alerts` (BFF) → `GET /v1/alerts` (fraud-service) →
`alloydb.RealClient.ListAlerts` → Postgres. The Next.js hook
`useAlerts` handles loading / error / simulated flags and re-fetches.

### Step 5 — Analyst triages

The analyst clicks an alert, sees the SHAP explainability panel
(`ShapExplainability.tsx`), opens the case, and either:
- closes it (`MarkResolved`),
- escalates to compliance (`POST /v1/compliance/sar/draft`).

### Step 6 — Compliance draft

`compliance-service` calls `compliance_llm` (Python), which uses
Gemini to draft the SAR narrative, then writes the immutable record
to BigQuery `audit_log` (append-only).

That's the whole loop.

---

## 7. The Go services in detail

### 7.1 `fraud-service` — the hot path

Responsibilities: score transactions, ingest transactions, list alerts.

Endpoints:
- `POST /v1/score` — given a transaction, return APPROVE/REVIEW/BLOCK
- `POST /v1/transactions` — ingest a transaction (for replay/testing)
- `GET /v1/alerts` — list alerts (newest first, optional `?customer_id`, `?limit`)
- `GET /healthz` — liveness

Key files:
- `main.go` — boot: load config, wire dependencies, start HTTP server
- `handler.go` — `handleScore`, `handleIngest`, `handleListAlerts`
- `decision.go` — score → decision mapping
- `scorer.go` — orchestrates feature lookup → Vertex predict → decision
- `enricher.go` — request enrichment (default values, validation)

Wiring (in `main.go`):
```go
var (
    predictorInst  vertex.Predictor       = &vertex.MockPredictor{ModelVersion: "mock-v1"}
    featureC       bigquery.FeatureClient = bigquery.MockClient{}
    alertStoreInst alloydb.Client         = alloydb.Select(ctx, cfg)  // ⭐ picks mock or pgx
    publisher      pubsub.Publisher       = pubsub.NewMockPublisher()
)
```

`alloydb.Select(ctx, cfg)` is the live/mock switch. If `ALLOYDB_DSN` is
set, it opens a pgx pool. If the pool can't connect, it logs to stderr
and returns the mock so the service still boots.

### 7.2 `risk-service` — the customer risk view

Endpoints:
- `GET /v1/risk/{customer_id}` — composite 0–1000 risk score
- `GET /v1/customers/{id}/transactions` — recent transactions (powers the customer dashboard)
- `POST /v1/risk/recompute` — force a recompute
- `GET /healthz`

The interesting file is the `riskstore` package — see section 9.

### 7.3 `chat-bff` — the LLM gateway

Endpoints:
- `POST /v1/chat` — `{ message, session_id? }` → `{ reply, session_id, action, redacted }`

What it does:
1. Looks up or creates a session in memory.
2. Redacts PII (SSN, card, email) using `pkg/dlp/redactor.go`.
3. Calls `gemini_service` (Python) over HTTP.
4. Returns the redacted reply.

Why a Go BFF for a Python LLM service? Three reasons:
- Session state lives in Go (single binary, easy to scale).
- PII redaction is enforced at the boundary — Python never sees raw PII.
- The Python service stays focused on Gemini + RAG.

### 7.4 `document-service` (Go BFF) + `document_service` (Python)

Go BFF: `POST /v1/documents/upload` (multipart) → GCS → call Python.
Python: `POST /extract` → Document AI; `POST /validate` → schema check.

Why split: Document AI has rich Python SDK support; the Go BFF handles
auth and GCS uploads.

### 7.5 `compliance-service` + `compliance_llm`

Go service: `POST /v1/compliance/sar/draft` (async, returns 202).
Python service: LLM-drafts the SAR narrative, writes to BigQuery audit_log.

### 7.6 `api-gateway`

A simple Go reverse proxy with JWT validation, rate limit, and routing.
Production uses Apigee; this is a local-dev convenience.

### 7.7 `alert-worker`

Background Pub/Sub subscriber. Reads `alerts-sub`, fans out to all
WebSocket clients (analyst console real-time view).

### 7.8 `db-bootstrap`

One-shot binary run by `scripts/db_bootstrap.sh`. Applies the alerts
schema and seeds 5 alerts into local Postgres. Idempotent.

### 7.9 Shared packages — the `services/pkg/` directory

These are the reason the Go services are small. Quick tour:

| Package | What it does |
|---|---|
| `config` | Loads env vars into a typed `*Config`. 12-factor. Includes `Load()`, `IsFlag()`, accessors for durations. |
| `httpapi` | HTTP server wrapper, middleware (logging, recovery, request-id, CORS), standard error envelope. |
| `logging` | Structured logger. `FromContext` returns a per-request logger with request_id pre-bound. |
| `tracing` | OpenTelemetry tracer setup. Exporter chosen by `OTEL_EXPORATOR` env (`none`/`otlp`/`console`). |
| `auth` | JWT verifier (Identity Platform format). |
| `domain` | All shared Go types: `Alert`, `AlertEnvelope`, `Transaction`, `CustomerTransactions`, `RiskScore`, `Document`. |
| `alloydb` | ⭐ Alerts store. `Client` interface, `MockClient`, `RealClient` (pgx), `Select(ctx, cfg)`, `IsMock(c)`. |
| `bigquery` | Feature client. Same mock/real pattern. |
| `riskstore` | ⭐ Transactions store. `Client` interface, `MockClient` (deterministic), `BigQueryClient` (raw HTTP), `Select(cfg)`, `IsMock(c)`. |
| `pubsub` | Publisher + subscriber. Mock + real. |
| `vertex` | Vertex AI predictor. Mock + real. |
| `dlp` | PII redactor used by `chat-bff`. |

The two ⭐ packages are the heart of the "live data" plan — see
section 12.

---

## 8. The Python services in detail

### 8.1 `gemini_service`

FastAPI app exposing:
- `POST /chat` — given `{ message, session_id? }`, return Gemini's reply
- `POST /complete` — non-chat completion

Powers RAG by reading from Vertex AI Search.

### 8.2 `document_service`

FastAPI app:
- `POST /extract` — runs Document AI on a GCS object, returns extracted fields
- `POST /validate` — schema check on the extracted JSON

### 8.3 `compliance_llm`

FastAPI app:
- `POST /draft_sar` — drafts a Suspicious Activity Report narrative
- `POST /draft_ctr` — drafts a Currency Transaction Report

### 8.4 `risk_batch`

Nightly batch job. Reads the day's transactions from BigQuery, computes
features, writes to the feature store, optionally retrains the fraud
XGBoost model (via Vertex AI Pipelines).

### 8.5 The `libs/` directory

These are the reusable Python building blocks. Each has its own tests
(47/47 pass):
- `finguard_common` — clients, config, logging, schemas
- `finguard_dlp` — PII redactor (SSN, card, email)
- `finguard_docai` — Document AI extractor + validators
- `finguard_evaluation` — metrics + SHAP
- `finguard_gemini` — chat, guardrails, RAG
- `finguard_vertex` — feature store, predictor, registry

---

## 9. The data layer

### 9.1 BigQuery (analytics, feature store, audit log)

- `finguard-dev.fin.customer_features_mv` — feature view for fraud scoring
- `finguard-dev.finguard_dev.transactions` — transactions for the customer dashboard
- `finguard-dev.finguard_dev.audit_log` — append-only compliance audit log
- Schemas in `terraform/modules/data/schemas/*.json` (BigQuery schemas) and `*.sql` (Postgres)

### 9.2 AlloyDB (OLTP) — the alerts store

- `finguard.alerts` table — one row per alert
- Schema in `terraform/modules/data/schemas/alerts.sql`

In local dev, we run Postgres:16 via docker-compose on port 5433.
AlloyDB is wire-compatible with Postgres, so `pgx` works against both.

### 9.3 Cloud Storage

- `finguard-documents` — uploaded documents
- `finguard-dlq` — Pub/Sub dead-letter queue

### 9.4 Vertex AI Feature Store (online + offline)

Online for low-latency feature lookups in `fraud-service`; offline for
batch training.

### 9.5 Memorystore (Redis)

Optional cache layer for hot features. Falls back to direct BQ on miss.

### 9.6 Pub/Sub DLQ

Every subscriber has a DLQ for messages that fail to process.

### 9.7 The `riskstore` package — a deep dive

This is the most interesting package because it has **two interchangeable
implementations** of the same interface.

**Interface (`client.go`):**
```go
type Client interface {
    ListCustomerTransactions(ctx context.Context, customerID string, limit int) (domain.CustomerTransactions, error)
}
```

**MockClient (`mock.go`):**
- Deterministic — same customerID always returns the same transactions.
- `seedMerchantPool` is a small catalog of 6 merchants (blue-bottle, whole-foods, uber, amazon, shell, netflix).
- The function rewrites each row's TxnID with a stable hash of `customerID:merchantID`, so different customers get different but stable IDs.
- Returns `simulated:true, source:"riskstore:MockClient"`.

**BigQueryClient (`bigquery.go`):**
- Talks to the goccy BigQuery emulator over raw HTTP (the emulator is auth-less).
- POSTs `{"query": "SELECT … FROM … WHERE customer_id = '<id>' …"}` to `/bigquery/v2/projects/{project}/queries`.
- The only user-controlled string is `customerID`. It is validated by `isSafeBQString` (only `[A-Za-z0-9._-]`, max 128 chars) **before** interpolation.
- Returns `simulated:false, source:"riskstore:BigQueryClient"`.

**Select (`select.go`):**
```go
func Select(cfg *config.Config) Client {
    if cfg != nil && cfg.BigQueryEmulatorHost != "" {
        return &BigQueryClient{
            Host:    cfg.BigQueryEmulatorHost,
            Project: cfg.RiskstoreBQProject,
            Dataset: cfg.RiskstoreBQDataset,
        }
    }
    return MockClient{}
}
```

**IsMock (in `select.go`):**
```go
func IsMock(c Client) bool {
    _, ok := c.(MockClient)
    return ok
}
```

`IsMock` is the single source of truth for the `simulated` flag in the
HTTP layer. If you add a new `Client` implementation, **you must update
`IsMock`** — otherwise the flag will be wrong.

### 9.8 The `alloydb` package — same pattern

`pg_pgx.go` is the live implementation. `Select(ctx, cfg)` opens a pgx
pool, falls back to `MockClient{}` on connect failure (with a stderr log
line). `IsMock(c)` is the type-assertion trick again.

---

## 10. The BFF (Next.js) layer

### 10.1 What is a BFF?

A **Backend-For-Frontend** is a thin server-side layer that sits between
the browser and the actual microservices. It exists to:
- Centralize auth, headers, retries.
- Aggregate calls (one browser request → multiple backend calls).
- Avoid CORS pain.
- Provide a stable API surface for the frontend team.

In our case, the BFF is **Next.js Route Handlers** in
`frontend/src/app/api/`. They proxy to the Go services.

### 10.2 The route handlers

| Route | Method | Forwards to |
|---|---|---|
| `/api/score` | POST | `fraud-service:8080/v1/score` |
| `/api/chat` | POST | `chat-bff:8082/v1/chat` |
| `/api/alerts` | GET | `fraud-service:8080/v1/alerts` |
| `/api/customers/[id]/transactions` | GET | `risk-service:8081/v1/customers/{id}/transactions` |
| `/api/healthz` | GET | aggregates all 10 services |

### 10.3 The `MOCK_*` safety net (in `lib/auth.ts`)

The frontend keeps a small in-process `MOCK_ALERTS` / `MOCK_TXNS`
fallback that **only fires when the BFF itself is unreachable**. It is
deliberately not triggered by `simulated:true` responses, because we
want the "Simulated data" badge to be visible to the user when the
backend is up but in mock mode (a development scenario).

### 10.4 The hook layer (`src/hooks/`)

- `useAlerts` — `loading / error / simulated` flags + refetch
- `useCustomerTransactions` — same, for transactions
- `useRiskScore` — for the risk gauge
- `useChat` — chat session state

These are the unit of reusability for components.

### 10.5 The component layer

- `analyst/AlertQueue.tsx` — table of alerts with loading/empty/error/simulated states
- `analyst/AlertCard.tsx` — per-row "Simulated data" badge
- `analyst/RiskGauge.tsx` — radial gauge for the 0–1000 score
- `analyst/ShapExplainability.tsx` — SHAP values for the score
- `analyst/CaseTimeline.tsx` — case events
- `customer/ChatWindow.tsx` — message list + input
- `customer/TransactionList.tsx` — table with loading/empty/error/simulated states
- `customer/DisputeForm.tsx` — open a dispute
- `layout/Sidebar.tsx`, `layout/Topbar.tsx`, `layout/RoleSwitcher.tsx`

### 10.6 UI primitives

`src/components/ui/` has shadcn-style primitives: Button, Card, Dialog,
Badge, Table. They wrap Radix UI + Tailwind.

---

## 11. Configuration and the `MOCK_MODE` contract

Every FinGuard service follows the same 12-factor config pattern. All
config comes from env vars. `config.Load()` in
`services/pkg/config/config.go` returns a typed `*Config`.

Key env vars:

| Var | Used by | Default | What it does |
|---|---|---|---|
| `SERVICE_NAME` | all | service-specific | Used in logs and metrics |
| `ENV` | all | `dev` | `dev`/`staging`/`prod` |
| `PORT` | all | 8080 | HTTP port |
| `GCP_PROJECT_ID` | all | `finguard-dev` | GCP project (required) |
| `GCP_REGION` | all | `asia-south1` | GCP region |
| `MOCK_MODE` | all | `true` | When true, GCP clients return mocks |
| `MOCK_VERTEX` | fraud-service | `true` | Mock the Vertex AI predictor |
| `FEATURE_FLAGS` | all | (none) | Comma-separated `k=v,k=v` flags |
| `ALLOYDB_DSN` | fraud-service | (unset) | When set, use pgx; otherwise mock |
| `BIGQUERY_EMULATOR_HOST` | risk-service | (unset) | When set, use BQ HTTP client; otherwise mock |
| `RISKSTORE_BQ_PROJECT` | risk-service | `finguard-dev` | Project for the BQ client |
| `RISKSTORE_BQ_DATASET` | risk-service | `finguard_dev` | Dataset for the BQ client |
| `SCORE_TIMEOUT` | fraud-service | `450ms` | Hard deadline for the score path |
| `BIGQUERY_TIMEOUT` | fraud-service | `200ms` | BQ lookup deadline |
| `VERTEX_TIMEOUT` | fraud-service | `250ms` | Vertex AI deadline |
| `VERTEX_FALLBACK` | fraud-service | `rule_based` | What to do if Vertex AI times out |

`MOCK_MODE=true` is the **safety net**: it means "you can boot without
any GCP credentials". Every GCP-touching code path checks this flag and
returns a mock implementation.

`ALLOYDB_DSN` and `BIGQUERY_EMULATOR_HOST` are **the live-data toggles**.
Set them and the service flips from mock to real, transparently.

---

## 12. The "live data" plan, in plain English

This is the most important recent change in the codebase. Before the
plan, the `simulated:false` claim in the BFF was a **lie** — every
response was actually coming from a mock. The plan made the claim honest.

### What the plan changed

We added a **second implementation** for the two stores that previously
only had mocks:
- **Alerts store** → `alloydb.RealClient` (pgx) + `alloydb.MockClient`
- **Transactions store** → `riskstore.BigQueryClient` (raw HTTP) + `riskstore.MockClient`

Each pair shares a single `Client` interface. The service calls
`Select(ctx, cfg)` to get the right one based on env vars. The HTTP
layer uses `IsMock(c)` to decide whether the envelope's `simulated`
flag is true or false.

### Why raw HTTP for BigQuery instead of the Go SDK

Three reasons:
1. **No SDK dependency.** The Go BQ SDK pulls in a lot; we wanted the
   service binary to stay small and boot fast.
2. **No auth for the emulator.** The goccy BQ emulator is auth-less;
   the SDK requires a token.
3. **Same code path for local and real.** When we point at real BQ, we
   add an `Authorization: Bearer …` header in one place.

### Why we kept the `customerID` interpolation safe

The only user-controlled string in the BQ query is `customerID`. It
goes into a single-quoted SQL literal. We validate it with
`isSafeBQString`:
- 1–128 chars
- only `[A-Za-z0-9._-]`

If a customer ID doesn't pass, the query is rejected before it hits BQ.
This is a single chokepoint — **do not string-escape at call sites**.

### How `restart_all.sh` wires it

`USE_LIVE_DATA=true` (default) makes `restart_all.sh`:
1. `docker compose up -d postgres` → Postgres on :5433
2. Install/start `bigquery-emulator` on :9050
3. Run `db-bootstrap` to apply schema + seed 5 alerts
4. Run `seed_local_http.py` to seed 50 transactions for `C-1001`
5. Export `ALLOYDB_DSN` and `BIGQUERY_EMULATOR_HOST` to every Go service
6. Start the 10 services

`USE_LIVE_DATA=false` skips steps 1–5 and runs the services in mock
mode only.

### How section 8 of `_test_bff.sh` proves it

After running `restart_all.sh`, `_test_bff.sh section 8`:
1. Probes `GET /v1/alerts` → expects `simulated:false, source:"alloydb.Postgres"`.
2. Probes `GET /v1/customers/C-1001/transactions` → expects `simulated:false, source:"riskstore:BigQueryClient"`, 50 rows.
3. Probes the BFF equivalents → expects `simulated:false` passthrough.
4. Checks the **shape** of the data (seeded IDs, source labels).

If any of these fail, the test fails loudly. This is what made the
`simulated:false` claim honest.

---

## 13. MLOps: how models are trained, deployed, and monitored

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

### 13.1 Pipelines (`ml/pipelines/`)

Each pipeline is a Kubeflow DSL Python file:
- `fraud_training_pipeline.py` — full retrain (validation → feature eng → train → evaluate → register → deploy)
- `fraud_continuous_training.py` — incremental retrain on the last N days
- `feature_engineering_pipeline.py` — rebuilds the feature store
- `docai_finetune_pipeline.py` — fine-tunes a custom Document AI processor

### 13.2 Components (`ml/pipelines/components/`)

- `data_validation.py` — TFDV-based schema validation
- `feature_eng.py` — feature transforms (one-hot, embed, scale)
- `train.py` — model training (XGBoost or autoencoder)
- `evaluate.py` — metrics (AUC, precision/recall, F1)
- `register.py` — registers the model in Vertex AI Model Registry
- `deploy.py` — deploys to an endpoint

### 13.3 Training entry points (`ml/training/`)

- `fraud_xgb/train.py` — XGBoost fraud model
- `fraud_autoencoder/train.py` — autoencoder for anomaly detection
- `docai_finetune/train.py` — custom Document AI processor

### 13.4 Deployment

- Vertex AI Endpoint for online prediction (`fraud-service` calls it)
- Batch prediction for nightly scoring (`risk_batch`)

### 13.5 Monitoring

Vertex AI Model Monitoring tracks:
- **Drift** — feature distribution shift
- **Skew** — training vs serving skew
- **Latency** — prediction latency
- **Error rate** — prediction errors

If drift or skew exceeds a threshold, an alert is published. The
`fraud-spike.md` and `model-drift.md` runbooks describe the response.

---

## 14. Security and compliance

### 14.1 Trust boundary (VPC-SC)

All GCP services sit inside a VPC Service Controls perimeter. Nothing
crosses the perimeter without an explicit access level.

### 14.2 Encryption

- **CMEK** (Customer-Managed Encryption Keys) via KMS for BigQuery, GCS, AlloyDB, Pub/Sub, Vertex AI.
- **In transit:** TLS everywhere.
- **At rest:** CMEK + default Google-managed encryption.

### 14.3 PII redaction

- `chat-bff` (Go) redacts SSN, card numbers, and emails **before**
  calling `gemini_service` (Python). The Python service never sees raw PII.
- The PII redactor has its own test suite (`finguard_dlp/tests/`).

### 14.4 Auth

- **Users:** IAP + Identity Platform. OAuth2 access tokens.
- **Service-to-service:** Workload Identity. No service account keys.

### 14.5 Audit

Cloud Audit Logs are streamed to a BigQuery `audit_log` table (append-only)
with 7-year retention. Every state change is recorded.

### 14.6 Network

- Hub-and-spoke VPC.
- Private Service Connect for Google APIs.
- Cloud NAT for egress.
- No public IPs in production.

### 14.7 Compliance filings

- `compliance-service` (Go) drafts SAR/CTR filings.
- `compliance_llm` (Python) uses Gemini to write the narrative.
- Every draft is logged to the immutable `audit_log` table.
- The narrative is reviewed by a human before submission to FinCEN.

---

## 15. Infrastructure as code

`terraform/` is a monorepo with three layers:

1. **Bootstrap** (`terraform/bootstrap/`) — runs once per org. Creates
   the project factory, billing account, and org-level resources.
2. **Modules** (`terraform/modules/`) — reusable per-resource modules.
   Each has its own README, variables, outputs, and versions.tf.
3. **Envs** (`terraform/envs/{dev,staging,prod}/`) — composes modules
   into a full environment. Each env has its own state.

The decoupling (ADR-0004) lets us change a module without touching an
env, and change an env without touching the bootstrap.

All 10 modules and all 3 envs validate clean with `terraform validate`.

---

## 16. CI/CD

### 16.1 Cloud Build (`deploy/cloudbuild/`)

- `go-service.yaml` — `go test`, `go vet`, build distroless image, push to Artifact Registry
- `python-service.yaml` — `pytest`, build, push
- `frontend-ci.yaml` — `npm test`, `npm run build`, build, push
- `ml-pipeline-run.yaml` — triggers a Vertex AI Pipeline
- `ml-train.yaml` — runs a training job
- `terraform-plan.yaml` / `terraform-apply.yaml` — `terraform plan`/`apply`
- `pr-checks.yaml` — runs on PRs

### 16.2 Cloud Deploy (`deploy/clouddeploy/`)

Progressive delivery: `dev` → `staging` → `prod`. Each service has its
own delivery pipeline (`fraud-service.yaml`, `risk-service.yaml`, etc.).

### 16.3 GitHub Actions (`deploy/github-actions/`)

Mirror the Cloud Build checks for PR validation. Four workflows:
`go-ci.yml`, `python-ci.yml`, `frontend-ci.yml`, `terraform-ci.yml`.

---

## 17. Testing strategy

| Layer | Tool | Where | Status |
|---|---|---|---|
| Go unit tests | `go test` | `services/pkg/...` and `services/cmd/...` | 18 / 18 pass |
| Go live tests | `go test` gated on `TEST_ALLOYDB_DSN` / `TEST_BIGQUERY_EMULATOR_HOST` | `services/pkg/{alloydb,riskstore}/*_test.go` | green online & offline |
| Python unit tests | `pytest` | `ml/libs/...` | 47 / 47 pass |
| Python service tests | `pytest` | `ml/services/...` | green |
| Frontend component tests | Vitest | `frontend/tests/components.test.tsx` | green |
| Frontend wiring tests | Vitest | `frontend/tests/wiring.test.tsx` | 7 / 7 pass |
| Terraform | `terraform validate` | `terraform/{modules,envs}/...` | clean |
| BFF E2E | bash | `_test_bff.sh` | 35×200 + 5×202 + 7×400 + 1×404 + 1×502; section 8 = 6 deep live-data assertions |
| Load tests | k6 | `scripts/load_test/k6_score.js` | run on demand |

The key insight: **live-data tests are gated on env vars** so CI stays
fast and hermetic, but a developer who has `restart_all.sh` running
can run the same tests against real (emulated) services.

---

## 18. Tradeoffs and "what we'd do differently next time"

### Tradeoffs we accepted

| Decision | Pro | Con |
|---|---|---|
| Polyglot (Go + Python + TS) | Right tool for each job | More toolchains |
| Raw HTTP for BQ instead of SDK | Small binary, no auth for emulator, same code for local + real | Hand-rolled BQ REST client; `isSafeBQString` is the chokepoint |
| pgx for local Postgres | Wire-compatible with AlloyDB; lightweight | Pool tuning is local-dev tuned |
| Service-level mock fallback (not crash) | Services always boot | Operators must read the boot log |
| Live tests gated on env vars | CI stays hermetic | Risk of regression on the live path |
| Frontend `MOCK_*` only on BFF unreachable | Avoids masking real outages | Real outage shows static data |
| Hard-coded port assignments in `restart_all.sh` | Works on macOS bash 3.2 | Add a new service → edit 4 places |

### Things we would change next time

1. **Terraform Cloud / remote state from day 1.** Local state is fine
   for a single dev; staging and prod want a real backend.
2. **CI workflow that runs both offline and live tests** in parallel
   (today it's the developer's responsibility).
3. **OpenAPI spec generated from the Go code** (we have one in
   `terraform/modules/apigateway/openapi.yaml` but it's not derived).
4. **Service mesh (Cloud Service Mesh or Istio)** for mTLS between
   services. We rely on VPC-SC + Workload Identity today.
5. **Retry/backoff in `riskstore.BigQueryClient.query`** — currently
   we fail fast.
6. **Real auth header for BQ** in `bigquery.go::query` — one line
   before pointing at real BigQuery.
7. **Connection-pool metrics** for the pgx pool — operators want to
   see live connections, idle, etc.
8. **A single `docker-compose.yml` for the Python services too.**
   Today you have to start them by hand.

---

## 19. Glossary

| Term | Meaning |
|---|---|
| **ADR** | Architecture Decision Record. A short markdown file explaining a decision. |
| **AlloyDB** | Google Cloud's wire-compatible Postgres. We use `pgx` against it. |
| **BFF** | Backend-For-Frontend. A thin server-side layer (here, Next.js Route Handlers) that sits between the browser and microservices. |
| **CMEK** | Customer-Managed Encryption Key. KMS keys you control. |
| **Dataflow** | Google Cloud's managed Apache Beam runner. |
| **DLP** | Data Loss Prevention. PII detection / redaction. |
| **Document AI** | Google Cloud's OCR + extraction service. |
| **DR** | Disaster Recovery. |
| **FR-N** | Functional Requirement N (from the PRD). |
| **Gemini** | Google Cloud's LLM. |
| **IAP** | Identity-Aware Proxy. |
| **KMS** | Key Management Service. |
| **Live data** | Real (not mock) data. In our case, real Postgres + BigQuery. |
| **Mock data** | Deterministic in-process data returned by `MockClient`s. |
| **PII** | Personally Identifiable Information. |
| **Pub/Sub** | Google Cloud's messaging service. |
| **RAG** | Retrieval-Augmented Generation. LLM with a search index. |
| **RPO / RTO** | Recovery Point Objective / Recovery Time Objective. |
| **SAR / CTR** | Suspicious Activity Report / Currency Transaction Report. |
| **SHAP** | SHapley Additive exPlanations. ML model explainability. |
| **SLO** | Service Level Objective. |
| **VPC-SC** | VPC Service Controls. A security perimeter. |
| **Workload Identity** | GCP's way of giving a service account to a GKE/Cloud Run workload without keys. |

---

*This document was written after the project was deleted from the local
machine and the code is preserved on GitHub at
`https://github.com/bikash-20/AI-BUILDER-CUP-2026.git`. Last updated
2026-10-15.*
