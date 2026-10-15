#!/usr/bin/env bash
# Print the health status of every FinGuard AI service.
set -u

# Plain variables for compatibility with bash 3.2 (macOS system bash).
PORT_FRONTEND=3000
PORT_FRAUD=8080
PORT_RISK=8081
PORT_CHAT=8082
PORT_DOC=8083
PORT_COMP=8084
PORT_GW=8085
PORT_PY_GEMINI=8090
PORT_PY_DOC=8091
PORT_PY_COMP=8092
PORT_POSTGRES=5433
PORT_BQ_EMULATOR=9050

check() {
  local name="$1" port="$2" path="$3"
  local code
  code=$(curl -s -m 2 -o /dev/null -w '%{http_code}' "http://127.0.0.1:$port$path" 2>/dev/null || echo "000")
  printf "%-22s :%-6s %-7s %s\n" "$name" "$port" "$code" "http://127.0.0.1:$port$path"
}

check_tcp() {
  local name="$1" port="$2"
  local ok
  if nc -z 127.0.0.1 "$port" 2>/dev/null; then
    ok="up"
  else
    ok="down"
  fi
  printf "%-22s :%-6s %-7s %s\n" "$name" "$port" "$ok" "tcp://127.0.0.1:$port"
}

check_http_root() {
  local name="$1" port="$2"
  # goccy BQ emulator returns 4xx on the root path; probe a real query
  # path instead. The "200" indicates the emulator accepted the query.
  local code
  code=$(curl -s -m 2 -o /dev/null -w '%{http_code}' \
    -X POST "http://127.0.0.1:$port/bigquery/v2/projects/finguard-dev/queries" \
    -H 'Content-Type: application/json' -d '{"query":"SELECT 1"}' 2>/dev/null || echo "000")
  if [[ "$code" == "200" ]]; then
    printf "%-22s :%-6s %-7s %s\n" "$name" "$port" "up" "http://127.0.0.1:$port/"
  else
    printf "%-22s :%-6s %-7s %s\n" "$name" "$port" "down" "http://127.0.0.1:$port/"
  fi
}

printf "%-22s %-7s %-7s %s\n" "service" "port" "code" "url"
printf -- "-%.0s" {1..70}; echo
check "frontend (BFF)"     "$PORT_FRONTEND"  "/api/healthz"
check "fraud-service"      "$PORT_FRAUD"     "/healthz"
check "risk-service"       "$PORT_RISK"      "/healthz"
check "chat-bff"           "$PORT_CHAT"      "/healthz"
check "document-service"   "$PORT_DOC"       "/healthz"
check "compliance-service" "$PORT_COMP"      "/healthz"
check "api-gateway"        "$PORT_GW"        "/healthz"
check "gemini_service"     "$PORT_PY_GEMINI" "/healthz"
check "document_service"   "$PORT_PY_DOC"    "/healthz"
check "compliance_llm"     "$PORT_PY_COMP"   "/healthz"

echo
printf "%-22s %-7s %-7s %s\n" "infra" "port" "state" "url"
printf -- "-%.0s" {1..70}; echo
check_tcp "postgres (docker)" "$PORT_POSTGRES"
check_http_root "bq_emulator"  "$PORT_BQ_EMULATOR"
