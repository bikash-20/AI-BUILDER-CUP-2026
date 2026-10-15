#!/usr/bin/env bash
# setup.sh — install the local development toolchain.
#
# This script installs the following tools via Homebrew (or apt):
#   - go (>= 1.22)
#   - python (>= 3.11)
#   - node (>= 20)
#   - terraform (>= 1.6)
#   - k6
#   - skaffold
#   - pre-commit
#   - gcloud (https://cloud.google.com/sdk/docs/install)
#
# It also sets up git pre-commit hooks and the `gcloud` config.

set -euo pipefail

# Detect platform.
case "$(uname -s)" in
  Darwin) INSTALL="brew";;
  Linux)  INSTALL="sudo apt-get";;
  *) echo "Unsupported platform: $(uname -s)"; exit 1;;
esac

# 1) Base packages
if [ "$INSTALL" = "brew" ]; then
  brew update
  brew install \
    go@1.22 \
    python@3.11 \
    node@20 \
    terraform \
    k6 \
    skaffold \
    pre-commit \
    google-cloud-cli \
    jq \
    yq
else
  sudo apt-get update
  sudo apt-get install -y \
    golang-go \
    python3.11 \
    nodejs \
    jq
  # terraform, k6, skaffold, gcloud require manual install on Linux.
  echo "Please install terraform, k6, skaffold, and gcloud manually."
fi

# 2) Go workspace
if [ ! -f services/go.sum ]; then
  (cd services && go mod tidy)
fi

# 2b) BigQuery emulator (goccy) — required for the live transactions path.
#     Skipped silently if already on PATH or already installed in $GOBIN.
if ! command -v bigquery-emulator >/dev/null 2>&1 && [ ! -x "${GOBIN:-$HOME/go/bin}/bigquery-emulator" ]; then
  echo "Installing bigquery-emulator (goccy)..."
  go install github.com/goccy/bigquery-emulator/cmd/bigquery-emulator@v0.6.6
fi

# 2c) Docker (used by docker-compose for the local Postgres backing the
#     alerts store). We only check for the binary; pulling/running
#     containers happens in scripts/restart_all.sh.
if ! command -v docker >/dev/null 2>&1; then
  echo "WARN: docker not found; live alerts (Postgres) path will be disabled." >&2
fi

# 3) Python virtualenvs
python3.11 -m venv .venv-ml
.venv-ml/bin/pip install -r ml/requirements.txt
python3.11 -m venv .venv-dataflow
.venv-dataflow/bin/pip install -r dataflow/requirements.txt

# 4) Node modules
(cd frontend && npm install)

# 5) Pre-commit hooks
if [ -f .pre-commit-config.yaml ]; then
  pre-commit install
fi

# 6) gcloud config (idempotent)
if ! gcloud config get-value project >/dev/null 2>&1; then
  gcloud config set project finguard-dev
fi

echo ""
echo "✓ Toolchain installed."
echo ""
echo "Next steps:"
echo "  source .venv-ml/bin/activate      # ml/  Python"
echo "  source .venv-dataflow/bin/activate # dataflow/ Python"
echo "  (cd services && go test ./...)    # Go services"
echo "  (cd frontend && npm run dev)      # Frontend"
echo "  (cd terraform/envs/dev && terraform init)"
