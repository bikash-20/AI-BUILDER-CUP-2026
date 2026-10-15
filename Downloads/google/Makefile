# FinGuard AI — top-level Makefile
# Re-exports component-level make targets so you can run them from the repo root.

SHELL := /usr/bin/env bash
ROOT  := $(shell pwd)

# ---- Help -----------------------------------------------------------------
.PHONY: help
help:
	@echo "FinGuard AI — top-level targets"
	@echo ""
	@echo "  make setup            Install toolchains (Go, Python, Node, Terraform)"
	@echo "  make lint             Lint everything"
	@echo "  make test             Run all tests"
	@echo "  make build            Build all services"
	@echo "  make fmt              Format Go, Python, Terraform"
	@echo "  make verify           Run go build, go test, terraform validate, npm run build"
	@echo ""
	@echo "  make go.* targets     Forwarded to ./services (go build, go test, ...)"
	@echo "  make py.* targets     Forwarded to ./ml"
	@echo "  make tf.* targets     Forwarded to ./terraform"
	@echo "  make fe.* targets     Forwarded to ./frontend"
	@echo ""

# ---- Setup ----------------------------------------------------------------
.PHONY: setup
setup:
	@./scripts/setup.sh

# ---- Aggregate targets ----------------------------------------------------
.PHONY: lint
lint:
	@echo ">> Linting Go...";     (cd services  && golangci-lint run ./... || true)
	@echo ">> Linting Python..."; (cd ml         && ruff check libs services training pipelines || true)
	@echo ">> Linting Frontend...";(cd frontend  && npm run lint || true)
	@echo ">> Terraform fmt check...";(cd terraform && terraform fmt -check -recursive)

.PHONY: fmt
fmt:
	@echo ">> gofmt...";     (cd services && gofmt -s -w .)
	@echo ">> black + ruff...";(cd ml && black . && ruff check --fix . || true)
	@echo ">> prettier..."; (cd frontend && npm run format || true)
	@echo ">> terraform fmt...";(cd terraform && terraform fmt -recursive)

.PHONY: test
test: go.test py.test fe.test

.PHONY: build
build: go.build py.build fe.build

.PHONY: verify
verify:
	@echo ">> go build..."; (cd services && go build ./...)
	@echo ">> go test...";  (cd services && go test ./...)
	@echo ">> terraform validate..."; \
	  for m in $(shell find terraform/modules -name "main.tf" -exec dirname {} \;); do \
	    echo "  - $$m"; \
	  done
	@echo ">> python imports..."; (cd ml && python -c "import finguard_common, finguard_vertex, finguard_docai, finguard_gemini, finguard_dlp, finguard_evaluation" || true)

# ---- Go -------------------------------------------------------------------
.PHONY: go.build
go.build:
	cd services && go build ./...

.PHONY: go.test
go.test:
	cd services && go test ./...

.PHONY: go.run-fraud
go.run-fraud:
	cd services/cmd/fraud-service && MOCK_MODE=true go run .

.PHONY: go.run-gateway
go.run-gateway:
	cd services/cmd/api-gateway && MOCK_MODE=true go run .

.PHONY: go.tidy
go.tidy:
	cd services && go mod tidy

# ---- Python ---------------------------------------------------------------
.PHONY: py.install
py.install:
	cd ml && pip install -r requirements.txt -e libs/finguard_common -e libs/finguard_vertex -e libs/finguard_docai -e libs/finguard_gemini -e libs/finguard_dlp -e libs/finguard_evaluation

.PHONY: py.test
py.test:
	cd ml && pytest libs/ services/ training/ pipelines/ dataflow/

.PHONY: py.build
py.build:
	cd ml && python -m compileall libs services training pipelines || true

.PHONY: py.run-document
py.run-document:
	cd ml/services/document_service && MOCK_MODE=true uvicorn app:app --reload --port 8080

.PHONY: py.run-gemini
py.run-gemini:
	cd ml/services/gemini_service && MOCK_MODE=true uvicorn app:app --reload --port 8081

# ---- Frontend -------------------------------------------------------------
.PHONY: fe.install
fe.install:
	cd frontend && npm install

.PHONY: fe.build
fe.build:
	cd frontend && npm run build

.PHONY: fe.test
fe.test:
	cd frontend && npm test

.PHONY: fe.dev
fe.dev:
	cd frontend && npm run dev

# ---- Terraform -----------------------------------------------------------
.PHONY: tf.init-dev
tf.init-dev:
	cd terraform/envs/dev && terraform init -backend=false

.PHONY: tf.validate
tf.validate:
	@set -e; for d in $$(find terraform/modules -mindepth 1 -maxdepth 2 -name "main.tf" -exec dirname {} \;); do \
	  echo ">> $$d"; \
	  (cd $$d && terraform init -backend=false >/dev/null && terraform validate); \
	done
	@set -e; for env in dev staging prod; do \
	  echo ">> envs/$$env"; \
	  (cd terraform/envs/$$env && terraform init -backend=false >/dev/null && terraform validate); \
	done

.PHONY: tf.fmt
tf.fmt:
	cd terraform && terraform fmt -recursive

.PHONY: tf.plan-dev
tf.plan-dev:
	cd terraform/envs/dev && terraform plan
