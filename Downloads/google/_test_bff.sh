#!/usr/bin/env bash
# End-to-end test for FinGuard AI.
#
# Tests:
#   1. Frontend BFF (Next.js :3000) -> Go services (fraud, risk, chat)
#   2. Direct Go service endpoints
#   3. api-gateway routing to all backends (and 502 on unknown paths)
#   4. Python AI services (gemini, document, compliance)
#   5. Compliance SAR/CTR + audit log
#   6. PII redaction through chat-bff
#   7. Validation + error handling
#   8. Live data path verification (only meaningful when the services
#      were started with USE_LIVE_DATA=true and Postgres/BQ-emulator
#      are reachable; otherwise this section degrades gracefully.)
#
# All services must already be running:
#   - frontend          :3000
#   - fraud-service     :8080
#   - risk-service      :8081
#   - chat-bff          :8082
#   - document-service  :8083
#   - compliance-service:8084
#   - api-gateway       :8085
#   - gemini_service    :8090
#   - document_service  :8091
#   - compliance_llm    :8092
#
# Exit code 0 on success. Each section is independent and tolerant of the
# next one failing.
set -u

BFF=http://127.0.0.1:3000
GW=http://127.0.0.1:8085

fraud=http://127.0.0.1:8080
risk=http://127.0.0.1:8081
chat=http://127.0.0.1:8082
docu=http://127.0.0.1:8083
comp=http://127.0.0.1:8084

py_gemini=http://127.0.0.1:8090
py_docu=http://127.0.0.1:8091
py_comp=http://127.0.0.1:8092

# ---------- helpers ----------
hr() { printf '\n--- %s ---\n' "$1"; }
hit() {
  local label="$1" method="$2" url="$3" body="${4:-}"
  if [[ -n "$body" ]]; then
    out=$(curl -s -w "\nCODE=%{http_code}" -X "$method" "$url" -H 'Content-Type: application/json' -d "$body" 2>&1)
  else
    out=$(curl -s -w "\nCODE=%{http_code}" -X "$method" "$url" 2>&1)
  fi
  code=$(printf '%s' "$out" | tail -1 | cut -d= -f2)
  body_out=$(printf '%s' "$out" | sed '$d')
  printf '[%s] %s -> %s\n%s\n\n' "$code" "$label" "$url" "$body_out"
}

# ---------- 1. BFF happy paths ----------
hr "1. BFF happy paths"
hit "BFF /api/score (medium txn)" POST "$BFF/api/score" \
  '{"txn_id":"T-1","customer_id":"C-1","amount":99.99,"currency":"USD","merchant_id":"M-100","timestamp":"2026-10-15T10:00:00Z","channel":"WEB"}'

hit "BFF /api/score (small txn)" POST "$BFF/api/score" \
  '{"txn_id":"T-2","customer_id":"C-2","amount":12.50,"currency":"USD","merchant_id":"M-100","timestamp":"2026-10-15T10:00:00Z","channel":"MOBILE"}'

hit "BFF /api/score (large txn)" POST "$BFF/api/score" \
  '{"txn_id":"T-3","customer_id":"C-3","amount":9999.99,"currency":"USD","merchant_id":"M-200","timestamp":"2026-10-15T10:00:00Z","channel":"WEB"}'

hit "BFF /api/risk/C-1" GET "$BFF/api/risk/C-1"
hit "BFF /api/chat" POST "$BFF/api/chat" '{"message":"Why was my card declined?","session_id":"s-1"}'

# ---------- 2. BFF validation + error handling ----------
hr "2. BFF validation + error handling"
hit "BFF /api/score (unknown field)" POST "$BFF/api/score" \
  '{"txn_id":"T-4","customer_id":"C-4","amount":50,"currency":"USD","event_time":"2026-10-15T10:00:00Z","source":"WEB"}'
hit "BFF /api/score (missing required)" POST "$BFF/api/score" \
  '{"txn_id":"T-5","customer_id":"C-5","amount":50,"currency":"USD","timestamp":"2026-10-15T10:00:00Z"}'
hit "BFF /api/score (bad JSON)" POST "$BFF/api/score" '{not json'
hit "BFF /api/score (empty body)" POST "$BFF/api/score" ''

# ---------- 3. PII redaction through chat-bff ----------
hr "3. PII redaction"
hit "BFF /api/chat (SSN)" POST "$BFF/api/chat" \
  '{"message":"My SSN is 123-45-6789, please help.","session_id":"s-2"}'
hit "BFF /api/chat (card no.)" POST "$BFF/api/chat" \
  '{"message":"Card 4111 1111 1111 1111 was charged twice.","session_id":"s-3"}'
hit "BFF /api/chat (email)" POST "$BFF/api/chat" \
  '{"message":"Email me at john.doe@example.com about it.","session_id":"s-4"}'

# ---------- 4. Direct Go services ----------
hr "4. Direct Go services"
hit "fraud-service /v1/score" POST "$fraud/v1/score" \
  '{"txn_id":"T-direct","customer_id":"C-direct","amount":250,"currency":"USD","merchant_id":"M-1","timestamp":"2026-10-15T10:00:00Z","channel":"WEB"}'
hit "risk-service /v1/risk/C-direct" GET "$risk/v1/risk/C-direct"
hit "chat-bff /v1/chat" POST "$chat/v1/chat" '{"message":"hi","session_id":"s-direct"}'
hit "document-service /v1/documents/upload" POST "$docu/v1/documents/upload" \
  '{"doc_type":"kyc","file_url":"gs://finguard-documents/test-id.pdf"}'
hit "document-service /v1/documents/D-42" GET "$docu/v1/documents/D-42"
hit "compliance-service /v1/compliance/sar/draft" POST "$comp/v1/compliance/sar/draft" \
  '{"case_id":"CASE-1"}'
hit "compliance-service /v1/compliance/ctr/draft" POST "$comp/v1/compliance/ctr/draft" \
  '{"case_id":"CASE-2"}'
hit "compliance-service /v1/compliance/audit?case_id=CASE-1" GET "$comp/v1/compliance/audit?case_id=CASE-1"

# ---------- 5. api-gateway routing ----------
hr "5. api-gateway routing"
hit "api-gateway /v1/score" POST "$GW/v1/score" \
  '{"txn_id":"T-gw","customer_id":"C-gw","amount":75,"currency":"USD","merchant_id":"M-1","timestamp":"2026-10-15T10:00:00Z","channel":"CARD"}'
hit "api-gateway /v1/risk/C-gw" GET "$GW/v1/risk/C-gw"
hit "api-gateway /v1/chat" POST "$GW/v1/chat" '{"message":"gw test","session_id":"s-gw"}'
hit "api-gateway /v1/documents/upload" POST "$GW/v1/documents/upload" \
  '{"doc_type":"kyc","file_url":"gs://test/x.pdf"}'
hit "api-gateway /v1/compliance/sar/draft" POST "$GW/v1/compliance/sar/draft" \
  '{"case_id":"CASE-GW"}'
hit "api-gateway /v1/nope/x (expect 502 no_backend)" GET "$GW/v1/nope/x"

# ---------- 6. Python AI services ----------
hr "6. Python AI services"
hit "gemini_service /chat" POST "$py_gemini/chat" \
  '{"message":"How do I dispute a charge?","session_id":"s-py"}'
hit "gemini_service /healthz" GET "$py_gemini/healthz"
hit "document_service /extract" POST "$py_docu/extract" \
  '{"gcs_uri":"gs://test-bucket/x.pdf","doc_type":"KYC"}'
hit "document_service /validate" POST "$py_docu/validate" \
  '{"doc_type":"KYC","entities":[{"name":"John Doe","dob":"1990-01-01"}]}'
hit "compliance_llm /draft_sar" POST "$py_comp/draft_sar" \
  '{"case_id":"CASE-PY","notes":"multiple large transfers"}'
hit "compliance_llm /draft_ctr" POST "$py_comp/draft_ctr" \
  '{"case_id":"CASE-CTR","transactions":[{"amount":15000,"currency":"USD"}]}'

# ---------- 7. Live data endpoints (alerts + customer transactions) ----------
hr "7. Live data endpoints"

# Direct fraud-service.
hit "fraud-service /v1/alerts" GET "$fraud/v1/alerts"
hit "fraud-service /v1/alerts?customer_id=C-1001" GET "$fraud/v1/alerts?customer_id=C-1001"
hit "fraud-service /v1/alerts?limit=2" GET "$fraud/v1/alerts?limit=2"
hit "fraud-service /v1/alerts?limit=-1 (expect 400)" GET "$fraud/v1/alerts?limit=-1"

# BFF proxy.
hit "BFF /api/alerts" GET "$BFF/api/alerts"
hit "BFF /api/alerts?customer_id=C-1001" GET "$BFF/api/alerts?customer_id=C-1001"
hit "BFF /api/alerts?limit=2" GET "$BFF/api/alerts?limit=2"

# Direct risk-service.
hit "risk-service /v1/customers/C-1001/transactions" GET "$risk/v1/customers/C-1001/transactions"
hit "risk-service /v1/customers/C-1001/transactions?limit=3" GET "$risk/v1/customers/C-1001/transactions?limit=3"
hit "risk-service /v1/customers/C-1001/transactions?limit=0 (expect 400)" GET "$risk/v1/customers/C-1001/transactions?limit=0"
hit "risk-service /v1/customers/C-1001/nope (expect 404)" GET "$risk/v1/customers/C-1001/nope"

# BFF proxy.
hit "BFF /api/customers/C-1001/transactions" GET "$BFF/api/customers/C-1001/transactions"
hit "BFF /api/customers/C-1001/transactions?limit=3" GET "$BFF/api/customers/C-1001/transactions?limit=3"
hit "BFF /api/customers/bad id/transactions (expect 400)" GET "$BFF/api/customers/bad%20id/transactions"

# api-gateway routing for the new paths.
hit "api-gateway /v1/alerts" GET "$GW/v1/alerts"
hit "api-gateway /v1/alerts?customer_id=C-1001" GET "$GW/v1/alerts?customer_id=C-1001"
hit "api-gateway /v1/customers/C-1001/transactions" GET "$GW/v1/customers/C-1001/transactions"

# ---------- 8. Live data path verification ----------
hr "8. Live data path verification"

# Section 8 only checks the live path when the service was started with
# USE_LIVE_DATA=true. Detect by reading the envelope `simulated` field on
# the live-data endpoints. If false, the live path is active and we run
# deep assertions. If true, we just print a notice and skip.
alerts_env=$(curl -s -m 3 "$fraud/v1/alerts")
alerts_simulated=$(printf '%s' "$alerts_env" | grep -oE '"simulated":(true|false)' | head -1 | cut -d: -f2)
txns_env=$(curl -s -m 3 "$risk/v1/customers/C-1001/transactions")
txns_simulated=$(printf '%s' "$txns_env" | grep -oE '"simulated":(true|false)' | head -1 | cut -d: -f2)

if [[ "$alerts_simulated" == "false" && "$txns_simulated" == "false" ]]; then
  echo "Live path ACTIVE (simulated=false on both endpoints). Running deep assertions."

  # Alerts: source must be alloydb.Postgres; rows must include the
  # bootstrapped ALT-seed-001..005 ids and at least one row for C-1001.
  alerts_src=$(printf '%s' "$alerts_env" | grep -oE '"source":"[^"]*"' | head -1 | cut -d'"' -f4)
  if [[ "$alerts_src" != "alloydb.Postgres" ]]; then
    echo "  FAIL: alerts source = $alerts_src, want alloydb.Postgres"
  else
    echo "  OK  alerts source = $alerts_src"
  fi
  if printf '%s' "$alerts_env" | grep -q '"alert_id":"ALT-seed-001"'; then
    echo "  OK  alerts contains seeded ALT-seed-001"
  else
    echo "  FAIL: alerts missing seeded ALT-seed-001"
  fi
  if printf '%s' "$alerts_env" | grep -q '"customer_id":"C-1001"'; then
    echo "  OK  alerts contains C-1001 row"
  else
    echo "  FAIL: alerts missing C-1001 row"
  fi

  # Transactions: source must be riskstore.BigQueryClient.
  txns_src=$(printf '%s' "$txns_env" | grep -oE '"source":"[^"]*"' | head -1 | cut -d'"' -f4)
  if [[ "$txns_src" != "riskstore:BigQueryClient" ]]; then
    echo "  FAIL: txns source = $txns_src, want riskstore:BigQueryClient"
  else
    echo "  OK  txns source = $txns_src"
  fi

  # BFF passthrough should also carry simulated=false.
  bff_alerts=$(curl -s -m 3 "$BFF/api/alerts")
  bff_alerts_sim=$(printf '%s' "$bff_alerts" | grep -oE '"simulated":(true|false)' | head -1 | cut -d: -f2)
  if [[ "$bff_alerts_sim" == "false" ]]; then
    echo "  OK  BFF /api/alerts simulated=false"
  else
    echo "  FAIL: BFF /api/alerts simulated=$bff_alerts_sim"
  fi
  bff_txns=$(curl -s -m 3 "$BFF/api/customers/C-1001/transactions")
  bff_txns_sim=$(printf '%s' "$bff_txns" | grep -oE '"simulated":(true|false)' | head -1 | cut -d: -f2)
  if [[ "$bff_txns_sim" == "false" ]]; then
    echo "  OK  BFF /api/customers/C-1001/transactions simulated=false"
  else
    echo "  FAIL: BFF /api/customers/C-1001/transactions simulated=$bff_txns_sim"
  fi
else
  echo "Live path INACTIVE (alerts.simulated=$alerts_simulated, txns.simulated=$txns_simulated)."
  echo "Section 8 is a no-op when services were started with USE_LIVE_DATA!=true."
  echo "To exercise the live path:"
  echo "  docker compose up -d postgres"
  echo "  bigquery-emulator -host 127.0.0.1 -port 9050 &"
  echo "  bash scripts/db_bootstrap.sh"
  echo "  BIGQUERY_EMULATOR_HOST=http://127.0.0.1:9050 bash scripts/seed_local.sh"
  echo "  USE_LIVE_DATA=true ./scripts/restart_all.sh"
  echo "  bash _test_bff.sh"
fi

hr "DONE"
