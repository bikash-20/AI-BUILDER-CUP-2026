#!/usr/bin/env bash
# Stop every FinGuard AI service started by restart_all.sh.
set -euo pipefail

PORTS=(3000 8080 8081 8082 8083 8084 8085 8090 8091 8092)

for port in "${PORTS[@]}"; do
  pids=$(lsof -nP -iTCP:"$port" -sTCP:LISTEN -t 2>/dev/null || true)
  if [[ -n "$pids" ]]; then
    echo "killing :$port -> $pids"
    kill -TERM $pids 2>/dev/null || true
    sleep 0.2
    kill -KILL $pids 2>/dev/null || true
  fi
done
echo "done"
