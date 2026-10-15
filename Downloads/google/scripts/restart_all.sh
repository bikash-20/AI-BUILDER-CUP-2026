#!/usr/bin/env bash
# Start every FinGuard AI service locally.
#
# Idempotent: kills any previous instance bound to the same port first.
# Run from anywhere; uses absolute paths to the repo under
# /Users/bikashtalukder/Downloads/google.
#
# After this script finishes, verify with:
#   /Users/bikashtalukder/Downloads/google/scripts/status.sh
#   /Users/bikashtalukder/Downloads/google/_test_bff.sh
#
# Live data path:
#   When USE_LIVE_DATA=true (the default), this script also brings up
#   local Postgres (via docker compose) and the goccy BigQuery emulator,
#   bootstraps them, and starts the Go services with ALLOYDB_DSN and
#   BIGQUERY_EMULATOR_HOST set so the responses carry simulated=false.
#   Set USE_LIVE_DATA=false to start only the services in mock mode.
set -eo pipefail

REPO="/Users/bikashtalukder/Downloads/google"
LOGS="$REPO/.run-logs"
mkdir -p "$LOGS"

USE_LIVE_DATA="${USE_LIVE_DATA:-true}"
GOBIN_DIR="${GOBIN:-$HOME/go/bin}"
BQ_EMULATOR_HOST="${BQ_EMULATOR_HOST:-http://127.0.0.1:9050}"
ALLOYDB_DSN="${ALLOYDB_DSN:-postgres://finguard:finguard@127.0.0.1:5433/finguard?sslmode=disable}"

# Port assignments. Keep in sync with RUNBOOK_LOCAL.md.
# Using plain variables for compatibility with bash 3.2 (macOS system bash)
# where `set -u` + associative array literal can trip over hyphens in keys.
GO_PORTS_FRAUD=8080
GO_PORTS_RISK=8081
GO_PORTS_CHAT=8082
GO_PORTS_DOC=8083
GO_PORTS_COMP=8084
GO_PORTS_GW=8085
PY_PORT_GEMINI=8090
PY_PORT_DOC=8091
PY_PORT_COMP=8092
FRONTEND_PORT=3000
PORT_POSTGRES=5433
PORT_BQ_EMULATOR=9050

# Track the order of ports we expect to come up so we can poll them.
EXPECTED_PORTS=(
  "$GO_PORTS_FRAUD" "$GO_PORTS_RISK" "$GO_PORTS_CHAT"
  "$GO_PORTS_DOC" "$GO_PORTS_COMP" "$GO_PORTS_GW"
  "$PY_PORT_GEMINI" "$PY_PORT_DOC" "$PY_PORT_COMP"
  "$FRONTEND_PORT"
)

kill_port() {
  local port="$1"
  local pids
  pids=$(lsof -nP -iTCP:"$port" -sTCP:LISTEN -t 2>/dev/null || true)
  if [[ -n "$pids" ]]; then
    echo "  killing :$port (pids=$pids)"
    kill -TERM $pids 2>/dev/null || true
    sleep 0.4
    kill -KILL $pids 2>/dev/null || true
  fi
}

start_go() {
  local svc="$1" port="$2"
  echo "[go] $svc on :$port"
  kill_port "$port"
  local extra_env=""
  if [[ "$USE_LIVE_DATA" == "true" ]]; then
    extra_env="ALLOYDB_DSN=${ALLOYDB_DSN} BIGQUERY_EMULATOR_HOST=${BQ_EMULATOR_HOST}"
  fi
  ( cd "$REPO/services" && \
    env \
      "GCP_PROJECT_ID=finguard-dev" \
      "GCP_REGION=asia-south1" \
      "MOCK_MODE=true" \
      "MOCK_VERTEX=true" \
      "FEATURE_FLAGS=auto_redact_pii=true" \
      "LOG_LEVEL=info" \
      "SERVICE_NAME=$svc" \
      "PORT=$port" \
      $extra_env \
      nohup go run "./cmd/$svc" \
      > "$LOGS/$svc.log" 2>&1 & )
}

start_py() {
  local rel="$1" port="$2"
  local dir="$REPO/$rel"
  local name
  name=$(basename "$rel")
  echo "[py] $name on :$port"
  kill_port "$port"
  ( cd "$dir" && \
    env \
      "PYTHONPATH=$REPO/ml/libs" \
      "PORT=$port" \
      "MOCK_MODE=true" \
      nohup python3 -m uvicorn main:app --host 127.0.0.1 --port "$port" \
      > "$LOGS/$name.log" 2>&1 & )
}

start_frontend() {
  echo "[fe] next dev on :$FRONTEND_PORT"
  kill_port "$FRONTEND_PORT"
  ( cd "$REPO/frontend" && \
    env \
      "FRAUD_SERVICE_URL=http://127.0.0.1:8080" \
      "RISK_SERVICE_URL=http://127.0.0.1:8081" \
      "CHAT_SERVICE_URL=http://127.0.0.1:8082" \
      "DOCUMENT_SERVICE_URL=http://127.0.0.1:8083" \
      "COMPLIANCE_SERVICE_URL=http://127.0.0.1:8084" \
      nohup npm run dev -- --port "$FRONTEND_PORT" \
      > "$LOGS/frontend.log" 2>&1 & )
}

bring_up_postgres() {
  if [[ "$USE_LIVE_DATA" != "true" ]]; then
    echo "[infra] USE_LIVE_DATA!=true; skipping Postgres"
    return 0
  fi
  if ! command -v docker >/dev/null 2>&1; then
    echo "[infra] docker not found; alerts store will fall back to MockClient"
    return 0
  fi
  echo "[infra] docker compose up -d postgres"
  ( cd "$REPO" && docker compose up -d postgres ) >> "$LOGS/postgres.log" 2>&1 || true
  echo -n "[infra] waiting for Postgres on :$PORT_POSTGRES "
  for i in $(seq 1 30); do
    if nc -z 127.0.0.1 "$PORT_POSTGRES" 2>/dev/null; then
      echo " ok (after ${i}s)"
      return 0
    fi
    echo -n "."
    sleep 1
  done
  echo " FAILED — alerts store will fall back to MockClient"
}

bring_up_bq_emulator() {
  if [[ "$USE_LIVE_DATA" != "true" ]]; then
    echo "[infra] USE_LIVE_DATA!=true; skipping BQ emulator"
    return 0
  fi
  kill_port "$PORT_BQ_EMULATOR"
  local bq_bin="$GOBIN_DIR/bigquery-emulator"
  if [[ ! -x "$bq_bin" ]]; then
    if command -v bigquery-emulator >/dev/null 2>&1; then
      bq_bin="$(command -v bigquery-emulator)"
    else
      echo "[infra] bigquery-emulator not found; transactions store will fall back to MockClient"
      echo "[infra]   install with: go install github.com/goccy/bigquery-emulator/cmd/bigquery-emulator@v0.6.6"
      return 0
    fi
  fi
  echo "[infra] starting bigquery-emulator on :$PORT_BQ_EMULATOR"
  nohup "$bq_bin" -host 127.0.0.1 -port "$PORT_BQ_EMULATOR" \
    > "$LOGS/bq-emulator.log" 2>&1 &
  echo -n "[infra] waiting for BQ emulator on :$PORT_BQ_EMULATOR "
  for i in $(seq 1 15); do
    # The goccy emulator returns 4xx on the root path but 200 on a real
    # query path. Probe the queries endpoint.
    code=$(curl -s -o /dev/null -m 1 -w '%{http_code}' \
      -X POST "http://127.0.0.1:$PORT_BQ_EMULATOR/bigquery/v2/projects/finguard-dev/queries" \
      -H 'Content-Type: application/json' -d '{"query":"SELECT 1"}' 2>/dev/null || echo "000")
    if [[ "$code" == "200" ]]; then
      echo " ok (after ${i}s)"
      return 0
    fi
    echo -n "."
    sleep 1
  done
  echo " FAILED — transactions store will fall back to MockClient"
}

bootstrap_stores() {
  if [[ "$USE_LIVE_DATA" != "true" ]]; then
    return 0
  fi
  # Postgres alerts
  if nc -z 127.0.0.1 "$PORT_POSTGRES" 2>/dev/null; then
    echo "[infra] bootstrapping Postgres alerts"
    DATABASE_URL="$ALLOYDB_DSN" bash "$REPO/scripts/db_bootstrap.sh" >> "$LOGS/postgres.log" 2>&1 || \
      echo "[infra]   alerts bootstrap failed; service will fall back to MockClient"
  fi
  # BQ emulator
  code=$(curl -s -o /dev/null -m 1 -w '%{http_code}' \
    -X POST "http://127.0.0.1:$PORT_BQ_EMULATOR/bigquery/v2/projects/finguard-dev/queries" \
    -H 'Content-Type: application/json' -d '{"query":"SELECT 1"}' 2>/dev/null || echo "000")
  if [[ "$code" == "200" ]]; then
    echo "[infra] seeding BigQuery emulator"
    if command -v bq >/dev/null 2>&1; then
      ( cd "$REPO" && BIGQUERY_EMULATOR_HOST="$BQ_EMULATOR_HOST" bash scripts/seed_local.sh ) \
        >> "$LOGS/bq-emulator.log" 2>&1 || \
        echo "[infra]   BQ seed failed; service will fall back to MockClient"
    else
      ( cd "$REPO" && BIGQUERY_EMULATOR_HOST="$BQ_EMULATOR_HOST" python3 scripts/seed_local_http.py ) \
        >> "$LOGS/bq-emulator.log" 2>&1 || \
        echo "[infra]   BQ seed failed; service will fall back to MockClient"
    fi
  fi
}

echo "== Stopping anything on the target ports =="
for port in "${EXPECTED_PORTS[@]}"; do
  kill_port "$port"
done

if [[ "$USE_LIVE_DATA" == "true" ]]; then
  echo "== Bringing up local data stores =="
  bring_up_postgres
  bring_up_bq_emulator
  bootstrap_stores
fi

echo "== Starting Go services =="
start_go fraud-service      "$GO_PORTS_FRAUD"
start_go risk-service       "$GO_PORTS_RISK"
start_go chat-bff           "$GO_PORTS_CHAT"
start_go document-service   "$GO_PORTS_DOC"
start_go compliance-service "$GO_PORTS_COMP"
# api-gateway needs FINGUARD_*_URL env vars so it can resolve the
# fraud/risk/chat/document/compliance backends in local mode.
echo "[go] api-gateway on :$GO_PORTS_GW"
kill_port "$GO_PORTS_GW"
( cd "$REPO/services" && \
  env \
    "GCP_PROJECT_ID=finguard-dev" \
    "GCP_REGION=asia-south1" \
    "MOCK_MODE=true" \
    "MOCK_VERTEX=true" \
    "FEATURE_FLAGS=auto_redact_pii=true" \
    "LOG_LEVEL=info" \
    "SERVICE_NAME=api-gateway" \
    "PORT=$GO_PORTS_GW" \
    "FINGUARD_FRAUD_URL=http://127.0.0.1:8080" \
    "FINGUARD_RISK_URL=http://127.0.0.1:8081" \
    "FINGUARD_CHAT_URL=http://127.0.0.1:8082" \
    "FINGUARD_DOCUMENT_URL=http://127.0.0.1:8083" \
    "FINGUARD_COMPLIANCE_URL=http://127.0.0.1:8084" \
    nohup go run ./cmd/api-gateway \
    > "$LOGS/api-gateway.log" 2>&1 & )

echo "== Starting Python services =="
start_py ml/services/gemini_service   "$PY_PORT_GEMINI"
start_py ml/services/document_service "$PY_PORT_DOC"
start_py ml/services/compliance_llm   "$PY_PORT_COMP"

echo "== Starting frontend =="
start_frontend

echo
echo "== Waiting for /healthz on each service =="
ok=true
for port in "${EXPECTED_PORTS[@]}"; do
  if [[ "$port" == "$FRONTEND_PORT" ]]; then
    path="/api/healthz"
  else
    path="/healthz"
  fi
  for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30; do
    if curl -sf -m 1 "http://127.0.0.1:$port$path" >/dev/null 2>&1; then
      printf "  :%-5s  up (after %2ds)\n" "$port" "$i"
      break
    fi
    sleep 1
    if [[ $i -eq 30 ]]; then
      printf "  :%-5s  FAILED (see %s)\n" "$port" "$LOGS"
      ok=false
    fi
  done
done

if [[ "$USE_LIVE_DATA" == "true" ]]; then
  echo
  echo "== Live data store status =="
  if nc -z 127.0.0.1 "$PORT_POSTGRES" 2>/dev/null; then
    printf "  :%-5s  up (Postgres / alloydb)\n" "$PORT_POSTGRES"
  else
    printf "  :%-5s  down (MockClient fallback)\n" "$PORT_POSTGRES"
  fi
  code=$(curl -s -o /dev/null -m 1 -w '%{http_code}' \
    -X POST "http://127.0.0.1:$PORT_BQ_EMULATOR/bigquery/v2/projects/finguard-dev/queries" \
    -H 'Content-Type: application/json' -d '{"query":"SELECT 1"}' 2>/dev/null || echo "000")
  if [[ "$code" == "200" ]]; then
    printf "  :%-5s  up (BQ emulator / riskstore)\n" "$PORT_BQ_EMULATOR"
  else
    printf "  :%-5s  down (MockClient fallback)\n" "$PORT_BQ_EMULATOR"
  fi
fi

$ok && echo "OK" || echo "Some services failed to start; tail the logs in $LOGS"