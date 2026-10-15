# CHANGES.md — Recent Changes, Tradeoffs, and State of the Codebase

> **Audience:** another AI agent picking up this codebase.
> Read this first. It tells you what landed, why, where the critical parts
> live, and the tradeoffs the team deliberately accepted.
> Date of last full verification: **2026-10-15**.

---

## TL;DR — what just shipped

Five-phase **Real Database Integration** plan is complete. The contract
`simulated: false` is now **honest**: when the env is wired to a real
Postgres (alerts) and the goccy BigQuery emulator (transactions), the
responses come from those services; otherwise the in-process mocks serve
the same shape and the envelope/row `simulated` flag is `true`.

All five phases of the plan landed:

1. Live store layer (`services/pkg/{alloydb,riskstore}`)
2. Service wiring (`fraud-service`, `risk-service`)
3. Local infra (`docker-compose.yml`, `db-bootstrap`, `setup.sh`, `restart_all.sh`, `status.sh`)
4. Tests + RUNBOOK + E2E section 8 (`pg_pgx_test.go`, `bigquery_test.go`, `RUNBOOK_LOCAL.md`, `_test_bff.sh`)
5. End-to-end verification (all assertions green)

---

## 1. The "live data" plan, phase by phase

### Phase 1 — Live store layer (no wiring change yet)

Files created:

- `services/pkg/riskstore/client.go` — `Client` interface: `ListCustomerTransactions(ctx, customerID, limit) (domain.CustomerTransactions, error)`
- `services/pkg/riskstore/mock.go` — `MockClient`, deterministic `seedMerchantPool`, `stableHash` (FNV-1a-lite)
- `services/pkg/riskstore/bigquery.go` — `BigQueryClient` that POSTs raw JSON to the goccy BigQuery emulator's REST query endpoint; uses `isSafeBQString` to validate the only interpolated parameter (`customerID`); `Simulated=false, Source="riskstore:BigQueryClient"`
- `services/pkg/riskstore/select.go` — `Select(cfg)` constructor; `IsMock(c)` helper used by the HTTP layer to flip the per-row flag
- `services/pkg/alloydb/pg_pgx.go` — pgx-based `RealClient` with `NewRealClient(ctx, dsn)` (3 s ping, MaxConns=10, MinConns=1, MaxConnIdleTime=5 min), `InsertAlert`, `MarkResolved`, `ListAlerts(filter)`; `Select(ctx, cfg)` constructor that logs to a swappable `stderr` interface on connect failure and falls back to `MockClient`
- `services/pkg/config/config.go` — added `AlloyDBDSN`, `BigQueryEmulatorHost`, `RiskstoreBQProject`, `RiskstoreBQDataset` (env-gated, defaults preserve mock behavior)

### Phase 2 — Service wiring

- `fraud-service/main.go` — `alloydb.Select(ctx, cfg)` + log line on whether alerts store is `Mock` or `Real`
- `fraud-service/handler.go` — `handleListAlerts` derives per-row and envelope `simulated`/`source` from `alloydb.IsMock(s.alerts)`
- `risk-service/main.go` — moved `seedMerchantPool` + `stableHash` into `riskstore/mock.go`; `customerSubtree` delegates to a wired `riskstore.Client`
- `risk-service/handler_test.go` — updated to drive the `riskstore.MockClient` directly

### Phase 3 — Local infra

- `docker-compose.yml` — `postgres:16` on `:5433` (avoids clashing with system Postgres on 5432), container `finguard-postgres`, healthcheck `pg_isready`, named volume `finguard-postgres-data`
- `services/cmd/db-bootstrap/main.go` + `scripts/db_bootstrap.sh` — apply `alerts` schema, seed 5 alerts
- `scripts/seed_local_http.py` — pure-Python BigQuery seeder (no `bq` CLI dependency); `restart_all.sh` uses it automatically when `bq` is absent
- `scripts/setup.sh` — installs `bigquery-emulator` v0.6.6 and checks Docker
- `scripts/restart_all.sh` — `USE_LIVE_DATA=true` (default) brings up Postgres + BQ emulator, bootstraps both, exports `ALLOYDB_DSN` + `BIGQUERY_EMULATOR_HOST` to the Go services, prints infra status
- `scripts/status.sh` — adds `postgres (docker)` and `bq_emulator` rows

### Phase 4 — Tests + RUNBOOK + E2E section 8

- `services/pkg/alloydb/pg_pgx_test.go` — integration tests gated on `TEST_ALLOYDB_DSN`
- `services/pkg/riskstore/bigquery_test.go` — integration tests gated on `TEST_BIGQUERY_EMULATOR_HOST`
- `_test_bff.sh` section 8 — auto-detects live path via envelope `simulated` flag, then runs **6 deep assertions** (source labels, seeded IDs, BFF passthrough)
- `RUNBOOK_LOCAL.md` — new "Live vs simulated" section with manual verification steps

### Phase 5 — Verified end-to-end

- Postgres on :5433 ✓
- BQ emulator on :9050 ✓
- `GET /v1/alerts` → `simulated:false, source:"alloydb.Postgres"` ✓
- `GET /v1/customers/C-1001/transactions` → `simulated:false, source:"riskstore:BigQueryClient"`, 50 rows ✓
- BFF passthrough: `/api/alerts` and `/api/customers/{id}/transactions` carry `simulated:false` ✓
- `_test_bff.sh` section 8: all 6 deep assertions pass ✓
- `go test ./...` — all packages pass (online + offline variants of live tests both green) ✓

### Mock-fallback guarantee

`USE_LIVE_DATA=false ./scripts/restart_all.sh` brings up only the 10
services in mock mode. Responses carry `simulated:true, source:"alloydb.MockClient" / "riskstore:MockClient"`.

The frontend's `MOCK_ALERTS`/`MOCK_TXNS` fallback in `lib/auth.ts` is
unchanged and remains the safety net **only** when the BFF itself is
unreachable (not when it returns `simulated:true`).

---

## 2. Critical parts (read these if you change the live path)

| File | Why it's critical | What to watch |
|---|---|---|
| `services/pkg/riskstore/bigquery.go` | Talks to BQ over raw HTTP; only place we interpolate a user-controlled string | `isSafeBQString` **must** validate `customerID` before it lands in a query; size cap 128 chars; only `[A-Za-z0-9._-]` allowed |
| `services/pkg/riskstore/select.go` | Selects mock vs live; `IsMock` is the single source of truth for the `simulated` flag | Don't add a new `Client` type without updating `IsMock` — the flag will silently stay `true` |
| `services/pkg/alloydb/pg_pgx.go` | pgx connection pool; connect failure silently falls back to mock | The fallback path writes a stderr log line; tests rely on that behavior for "online vs offline" gating |
| `services/cmd/fraud-service/main.go` | Wires the live alerts store; reads `ALLOYDB_DSN` at boot | The log line at boot is the operator's only signal that the store is live |
| `services/cmd/risk-service/main.go` | Wires the live transactions store; reads `BIGQUERY_EMULATOR_HOST` at boot | Same: boot log is the only signal |
| `scripts/restart_all.sh` | Orchestrates Postgres + BQ emulator + 10 services; env wiring | Hard-coded port assignments; keep in sync with `RUNBOOK_LOCAL.md` |
| `docker-compose.yml` | Postgres :5433 | Don't switch to :5432 — clashes with system Postgres |
| `scripts/seed_local_http.py` | Pure-Python BQ seeder (no `bq` CLI) | If BQ schema changes, this needs to follow; queries must match `services/pkg/riskstore/bigquery.go`'s column expectations |
| `frontend/src/lib/auth.ts` | `MOCK_ALERTS` / `MOCK_TXNS` fallback; only fires when BFF is unreachable | Do **not** add a `simulated:true` short-circuit here — that hides real outages |

---

## 3. Tradeoffs (deliberately accepted)

| Decision | Pro | Con |
|---|---|---|
| goccy BigQuery emulator over `cloud.google.com/go/bigquery` SDK | No SDK dependency, no auth, fast boot, works offline | We hand-rolled a tiny BQ REST client; the `customerID` is the only interpolated parameter and **must** pass `isSafeBQString` |
| Raw `net/http` instead of a BQ SDK | Same code path will work against real BQ (add an `Authorization: Bearer …` header in one place) | We don't get streaming, parameterized queries, or query-cache hints; the emulator doesn't support `useLegacySql` or `parameterMode` |
| `pgx` for local Postgres (instead of full AlloyDB driver) | AlloyDB is wire-compatible with Postgres; lightweight; existing pg-mock tests still apply | Connection-pool tuning is local-dev tuned, not prod |
| Service-level mock fallback (instead of crash on missing infra) | Services still boot; `simulated` flag honestly reports the situation | Operators must read the boot log to know why data is simulated — this is a known UX cost |
| Live-data tests gated on env vars (`TEST_ALLOYDB_DSN`, `TEST_BIGQUERY_EMULATOR_HOST`) | CI stays fast & hermetic; offline developers aren't broken | Risk of regression on the live path if you only run offline tests; run `restart_all.sh` first |
| Frontend fallback only fires on BFF unreachable, not on `simulated:true` | Avoids masking real outages; users see the honest "Simulated data" badge from the envelope | A real outage shows static fallback data; considered a banner but adds noise |
| `restart_all.sh` hard-codes port assignments in plain bash variables | Compatible with bash 3.2 (macOS system bash) where associative arrays with hyphens trip up `set -u` | Adding a new service means editing 4 places (variable, kill_port call, start_go call, EXPECTED_PORTS) |
| Moved `seedMerchantPool` + `stableHash` from `risk-service/main.go` to `riskstore/mock.go` | Both mock and live paths share the data shape; no duplication | The mock no longer lives next to the service that uses it; if you change the row shape, change the shared `domain.Transaction` |

### Known limitations (not bugs, not yet)

- goccy BigQuery emulator v0.8.1 hits an `http2.TrailerPrefix` build error on this Go toolchain → we pin v0.6.6. When the upstream fix lands, bump and re-run the live tests.
- `riskstore.BigQueryClient` does not retry on 5xx; local dev doesn't need it, prod will.
- No connection-level auth header is sent — fine for the local emulator, must be added before pointing at real BigQuery.

---

## 4. What was done so far (chronological-ish)

- Initial scaffold: monorepo with `frontend/`, `services/`, `ml/`, `dataflow/`, `terraform/`, `deploy/`, `docs/`, `scripts/`. All Terraform modules/envs validate clean. All Go services compile. Python libs import. Frontend pages all 200.
- Service contracts: `services/pkg/domain/{alert,transaction}.go` carry `Simulated`; `CustomerTransactions` envelope defined.
- BFF: `frontend/src/app/api/{alerts,customers/[id]/transactions,healthz}/route.ts` proxy Go services.
- Hooks: `useAlerts`, `useCustomerTransactions` — loading/error/simulated flags.
- Components: `AlertQueue`, `AlertCard`, `TransactionList` — loading/empty/error/simulated states.
- Tests: `frontend/tests/wiring.test.tsx` (7/7), Vitest components (7/7).
- **`Real Database Integration` (5 phases, all complete):** see section 1.

---

## 5. Background commands still being monitored

Two background commands from the implementation work are historical artifacts:

- `bfin0vgjm` (failed) — first attempt at `bigquery-emulator` v0.8.1; hit the `http2.TrailerPrefix` build error. We already knew about that and pinned to v0.6.6.
- `bke73ra9e` (completed) — the v0.6.6 install that actually succeeded. This is the binary currently running on :9050 and serving the live path.

Both are non-actionable. The current state is unchanged:
- `bigquery-emulator` v0.6.6 installed at `/Users/bikashtalukder/go/bin/bigquery-emulator` and running on :9050.
- `docker compose up -d postgres` brought up `finguard-postgres` on :5433.
- Both live data endpoints return `simulated:false` with the correct source labels.
- All Go tests pass; `_test_bff.sh` section 8 deep assertions all pass.

---

## 6. How to reproduce the green state

```bash
cd /Users/bikashtalukder/Downloads/google
./scripts/restart_all.sh        # USE_LIVE_DATA=true (default)
./scripts/status.sh             # 12 components healthy
./_test_bff.sh                  # full E2E; section 8 runs deep live-data assertions
cd services && go test ./...    # 18/18 pass
cd ../ml && pytest libs/        # 47/47 pass
```

Mock-only (no Postgres, no BQ emulator):

```bash
USE_LIVE_DATA=false ./scripts/restart_all.sh
./_test_bff.sh                  # envelope carries simulated:true
```

---

## 7. If you change the live path — checklist

- [ ] If you add a column to `alerts`, update both the `INSERT` in `pg_pgx.go` and the bootstrap seeder in `services/cmd/db-bootstrap/`.
- [ ] If you add a column to `transactions` (BQ), update both the `BigQueryClient` decoder in `bigquery.go` and `scripts/seed_local_http.py`.
- [ ] If you add a new `riskstore.Client` implementation, **update `IsMock`** — otherwise the envelope's `simulated` flag will be wrong.
- [ ] If you change how `customerID` is validated, do it in `isSafeBQString` (single chokepoint) — do **not** string-escape at the call site.
- [ ] If you change the boot log line format, grep for it in `RUNBOOK_LOCAL.md` and `_test_bff.sh` section 8.
- [ ] If you change the port of any service, update `restart_all.sh`, `RUNBOOK_LOCAL.md`, and `_test_bff.sh`.
- [ ] If you change the env var names (`ALLOYDB_DSN`, `BIGQUERY_EMULATOR_HOST`, …), grep both `services/pkg/config/config.go` and `scripts/restart_all.sh` first.

---

## 8. Out of scope / not yet done

- Real AlloyDB driver (we use `pgx` against Postgres — wire-compatible).
- Real BigQuery auth header (we hit the local emulator; for prod, add `Authorization: Bearer …` in `bigquery.go::query`).
- v0.8.1+ of `bigquery-emulator` (pinned to v0.6.6 due to upstream `http2.TrailerPrefix` build error).
- Retry/backoff on the BQ HTTP path.
- Production deploy (terraform validates; we haven't run `terraform apply`).

---

*Last updated: 2026-10-15 — verified by a full `_test_bff.sh` + `go test ./...` + `pytest libs/` run.*
