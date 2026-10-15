#!/usr/bin/env bash
# db_bootstrap.sh — apply the alerts Postgres schema + seed data.
#
# Requires:
#   - The finguard-postgres container to be up (port 5433).
#   - go in PATH.
#
# Reads DATABASE_URL from the environment or uses the default pointing
# at the local docker-compose container.

set -euo pipefail

REPO="/Users/bikashtalukder/Downloads/google"
export DATABASE_URL="${DATABASE_URL:-postgres://finguard:finguard@127.0.0.1:5433/finguard?sslmode=disable}"

# Sanity-check the port first so we surface a clear error if Postgres isn't up.
if ! nc -z 127.0.0.1 5433 2>/dev/null; then
  echo "Postgres not reachable on 127.0.0.1:5433 — run:"
  echo "  docker compose up -d postgres"
  exit 1
fi

cd "$REPO/services"
go run ./cmd/db-bootstrap