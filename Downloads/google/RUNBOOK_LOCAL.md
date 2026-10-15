# FinGuard AI — Local Run Status

This document is a snapshot of the local end-to-end run completed on **2026-10-15**.

> **Live data path verified.** With `ALLOYDB_DSN` and `BIGQUERY_EMULATOR_HOST` set,
> `GET /v1/alerts` returns `simulated:false, source:"alloydb.Postgres"` against the
> local Postgres container, and `GET /v1/customers/{id}/transactions` returns
> `simulated:false, source:"riskstore:BigQueryClient"` against the goccy BigQuery
> emulator. Section 8 of `_test_bff.sh` asserts this end-to-end.

## What is running

| Component                       | Port | Process         | Health |
| ------------------------------- | ---: | --------------- | -----: |
| Next.js frontend (BFF)          | 3000 | node (Next dev) |    200 |
| fraud-service (Go)              | 8080 | go run          |    200 |
| risk-service (Go)               | 8081 | go run          |    200 |
| chat-bff (Go)                   | 8082 | go run          |    200 |
| document-service (Go)           | 8083 | go run          |    200 |
| compliance-service (Go)         | 8084 | go run          |    200 |
| api-gateway (Go)                | 8085 | go run          |    200 |
| gemini_service (Python / FastAPI)| 8090 | uvicorn         |    200 |
| document_service (Python)       | 8091 | uvicorn         |    200 |
| compliance_llm (Python)         | 8092 | uvicorn         |    200 |
| Postgres (docker, alerts store) | 5433 | docker compose  |   tcp  |
| BigQuery emulator (goccy)       | 9050 | `bigquery-emulator` | http  |

## Validation results

| Layer                         | Result |
| ----------------------------- | ------ |
| Go services compile (7)       | clean  |
| `go vet ./...`                | clean  |
| Go test packages              | 18 / 18 pass (17 OK + 1 with no tests) |
| Python libs import (`libs/`)  | OK     |
| Python tests (`pytest libs/`) | 47 / 47 pass |
| Terraform modules (10)        | all `validate` clean |
| Terraform envs (dev/stg/prod) | all `validate` clean |
| Frontend pages (`/`, `/analyst`, `/customer`, `/customer/chat`, `/analyst/cases/:id`) | all 200 |
| Frontend Vitest (`tests/components.test.tsx`, `tests/wiring.test.tsx`) | 7 / 7 pass |

## End-to-end BFF flow

`_test_bff.sh` exercises the full stack. Latest run:

| Status | Count | Category |
| -----: | ----: | -------- |
|    200 |    35 | happy paths (score, risk, chat, document get, audit, AI services, gateway, alerts, customer transactions) |
|    202 |     5 | async accepted (upload, SAR/CTR drafts) |
|    400 |     7 | validation (unknown field, missing required, empty body, bad JSON, bad limit, invalid customer id) |
|    404 |     1 | unknown sub-resource (`/v1/customers/C-1001/nope`) |
|    502 |     1 | api-gateway unknown path (`/v1/nope/x` -> no_backend) |

No 500 responses.

## Live data endpoints (new in this run)

Two new read paths surface seeded mock data through the same BFF conventions as the existing endpoints. Both are clearly identified as simulated via per-row and envelope `simulated`/`source` fields.

| Method | Path                                           | Backend     | Returns                                              |
| ------ | ---------------------------------------------- | ----------- | ---------------------------------------------------- |
| GET    | `/v1/alerts`                                   | fraud-svc   | Alerts (most-recent first) with `?customer_id`, `?limit` |
| GET    | `/v1/customers/{id}/transactions`              | risk-svc    | Transactions for `{id}` with `?limit`                  |
| GET    | `/api/alerts`                                  | BFF proxy   | Forwards to `/v1/alerts`                                |
| GET    | `/api/customers/{id}/transactions`             | BFF proxy   | Forwards to `/v1/customers/{id}/transactions`         |

The customer dashboard (`/customer`) and analyst console (`/analyst`) now read from these endpoints. The frontend renders a "Simulated data" badge whenever the envelope or row carries `simulated: true`, and falls back to embedded fallback data only when the BFF is unreachable.

## Live vs simulated

The `simulated` field is **derived from which client served the request**, not a constant. The wiring is:

| Env var                       | When set                                | When unset                  |
| ----------------------------- | --------------------------------------- | --------------------------- |
| `ALLOYDB_DSN`                 | `alloydb.RealClient` (pgx → Postgres)   | `alloydb.MockClient`        |
| `BIGQUERY_EMULATOR_HOST`      | `riskstore.BigQueryClient` (HTTP → goccy BQ emulator) | `riskstore.MockClient` |

The envelope and per-row `simulated` flag flips accordingly:

- `simulated=false`, `source="alloydb.Postgres"` — when `ALLOYDB_DSN` is set and the connection succeeded.
- `simulated=false`, `source="riskstore:BigQueryClient"` — when `BIGQUERY_EMULATOR_HOST` is set and the emulator answered.
- `simulated=true`, `source="alloydb.MockClient"` / `source="riskstore:MockClient"` — fallback path.

The local-dev defaults are set by `scripts/restart_all.sh` when run with `USE_LIVE_DATA=true` (the default). To run in **mock-only** mode (no Postgres, no BQ emulator):

```bash
USE_LIVE_DATA=false ./scripts/restart_all.sh
```

Manual verification of the live path:

```bash
# 1. Direct probe (should now be simulated=false)
curl -s http://127.0.0.1:8080/v1/alerts | jq '.simulated, .source'
# expect: false, "alloydb.Postgres"

curl -s http://127.0.0.1:8081/v1/customers/C-1001/transactions | jq '.simulated, .source'
# expect: false, "riskstore:BigQueryClient"

# 2. Frontend smoke
curl -s http://127.0.0.1:3000/api/alerts | jq '.simulated'
# expect: false

# 3. Full E2E (section 8 will run deep assertions when live data is on)
bash _test_bff.sh
```

## Data protection (PII redaction)

`chat-bff` redacts sensitive substrings before forwarding to the LLM:

| Input                               | Redacted reply                                                 |
| ----------------------------------- | -------------------------------------------------------------- |
| `My SSN is 123-45-6789...`          | `My SSN is [REDACTED_SSN]...` + action `verify_identity`       |
| `Card 4111 1111 1111 1111...`       | `Card [REDACTED_CARD]...` + action `verify_identity`           |
| `Email me at john.doe@example.com`  | `Email me at [REDACTED_EMAIL]...` + action `verify_identity`   |

## Key files touched during this run

- `services/pkg/config/config.go` — fixed `parseBool`, project-ID check
- `services/pkg/auth/verifier.go` — removed duplicate `Subject` shadowing
- `services/pkg/logging/logger.go` — `FromContext` returns `*Logger`
- `services/pkg/alloydb/pg.go` — added `ListAlerts(ctx, ListAlertsFilter)` to `Client`
- `services/pkg/domain/transaction.go` — added `Simulated` flag and `CustomerTransactions` envelope
- `services/pkg/domain/alert.go` — added `Simulated` flag
- `services/cmd/api-gateway/main.go` — `ErrorHandler` for unknown backend, routes `/v1/alerts` and `/v1/customers`
- `services/cmd/chat-bff/main.go` — added `SessionID` to response
- `services/cmd/document-service/main.go` — fixed status constant
- `services/cmd/fraud-service/handler.go` + `main.go` — added `GET /v1/alerts`
- `services/cmd/fraud-service/handler_test.go` — added alerts handler tests
- `services/cmd/risk-service/main.go` + `handler_test.go` — added `GET /v1/customers/{id}/transactions`
- `services/pkg/alloydb/pg_test.go` — added `ListAlerts` tests
- `frontend/src/app/api/score/route.ts` — return 400 on empty/bad body
- `frontend/src/app/api/chat/route.ts` — same defensive parsing
- `frontend/src/app/api/healthz/route.ts` — new BFF health aggregator
- `frontend/src/app/api/alerts/route.ts` — new BFF proxy for `/v1/alerts`
- `frontend/src/app/api/customers/[id]/transactions/route.ts` — new BFF proxy for customer transactions
- `frontend/src/hooks/useAlerts.ts` — loading / error / simulated flags
- `frontend/src/hooks/useCustomerTransactions.ts` — new hook for customer transactions
- `frontend/src/components/analyst/AlertQueue.tsx` — loading / empty / error / simulated states
- `frontend/src/components/analyst/AlertCard.tsx` — per-row simulated badge
- `frontend/src/components/customer/TransactionList.tsx` — fetches via hook; loading / empty / error / simulated states
- `frontend/src/app/customer/page.tsx` — passes `customerId` to `TransactionList`
- `frontend/src/lib/types.ts` + `auth.ts` — `Alert`, `Transaction`, `AlertsEnvelope`, `CustomerTransactions`, `listCustomerTransactions` API
- `frontend/tests/wiring.test.tsx` — new wiring tests for `TransactionList` (happy / loading / fallback / 502)
- `terraform/modules/organization/main.tf` — `list_policy.allow/deny`
- `terraform/modules/network/main.tf` — `secondary_ip_range` (singular)
- `terraform/modules/security/main.tf` — `versioned_expr` as attr
- `terraform/modules/data/main.tf` — removed destroy-time `local-exec`
- `terraform/modules/vertex_ai/main.tf` — switched to GA provider, removed model monitoring resource (provider doesn't support it yet)
- `terraform/modules/monitoring/main.tf` — `goal`, `rolling_period_days`, set-typed `method/location/version`
- `terraform/modules/alloydb/main.tf` — `recovery_window_days`, removed `deletion_protection` from cluster
- `terraform/envs/{staging,prod}/main.tf` — `pubsub_key_id` (was `pubsub_kms_key_id`)

## How to run again

```bash
# Bring up everything (idempotent). With USE_LIVE_DATA=true (default)
# this also brings up Postgres (docker compose) and the goccy BQ
# emulator, bootstraps both, and starts the Go services with
# ALLOYDB_DSN + BIGQUERY_EMULATOR_HOST set so responses carry
# simulated=false.
./scripts/restart_all.sh

# Mock-only mode (no Postgres, no BQ emulator)
USE_LIVE_DATA=false ./scripts/restart_all.sh

# Smoke check
./scripts/status.sh

# Full E2E (section 8 runs deep assertions when live data is active)
./_test_bff.sh
```

For an explicit walk-through:

```bash
# Terminal 1: backend services
cd services
for svc in fraud-service risk-service chat-bff document-service compliance-service api-gateway; do
  (cd cmd/$svc && SERVICE_NAME=$svc PORT=$((8080 + $(ls cmd | grep -n $svc | cut -d: -f1) - 1)) \
    GCP_PROJECT_ID=finguard-dev MOCK_MODE=true go run .) &
done

# Terminal 2: Python services
cd ml
for svc in services/gemini_service services/document_service services/compliance_llm; do
  (cd $svc && PYTHONPATH=../../libs PORT=$((8090 + RANDOM % 100)) uvicorn main:app --port ...) &
done

# Terminal 3: frontend
cd frontend && npm run dev

# Verify
./_test_bff.sh
```

The full E2E script is `_test_bff.sh` at the repo root.
