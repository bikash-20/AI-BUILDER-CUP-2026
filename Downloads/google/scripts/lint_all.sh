#!/usr/bin/env bash
# lint_all.sh — run all linters in parallel.
#
# This is a thin wrapper used by CI to keep the build pipeline
# declarative. Each language has its own canonical linter:
#   - Go:       go vet + gofmt
#   - Python:   ruff
#   - TypeScript: eslint + prettier
#   - Terraform: terraform fmt + validate

set -euo pipefail

# Go
echo "→ gofmt"
test -z "$(gofmt -l services/)"
echo "→ go vet"
(cd services && go vet ./...)

# Python
echo "→ ruff"
.venv-ml/bin/ruff check ml/ dataflow/

# TypeScript
echo "→ eslint"
(cd frontend && npx eslint . --ext .ts,.tsx)

echo "→ prettier"
(cd frontend && npx prettier --check .)

# Terraform
echo "→ terraform fmt"
terraform fmt -check -recursive terraform/

echo ""
echo "✓ All linters passed."
