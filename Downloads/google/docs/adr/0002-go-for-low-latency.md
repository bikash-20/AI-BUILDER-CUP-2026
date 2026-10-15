# ADR-0002 — Go for Latency-Critical Cloud Run Services

- **Status:** Accepted
- **Date:** 2026-10-15

## Context

`fraud-service` and `risk-service` need P95 < 500 ms. They orchestrate calls to Vertex AI (typical 50–150 ms), BigQuery feature lookups (50–200 ms cached, 200–500 ms cold), and AlloyDB writes. Container cold start and runtime overhead matter.

## Decision

Use **Go 1.22+** for the hot-path Cloud Run services: `fraud-service`, `risk-service`, `api-gateway`, `chat-bff`, `document-service` (Go BFF), `compliance-service`, `alert-worker`.

## Rationale

- Sub-100 ms cold start in Cloud Run
- Goroutine model fits fan-out to BigQuery + Vertex AI + AlloyDB
- Single static binary → simple distroless container
- Strong OpenTelemetry, gRPC, and Google Cloud client libraries

## Consequences

- Some logic (e.g., ML-specific transformations) lives in Python services called over HTTP; latency budget must account for the hop.
- Team must be fluent in Go — onboarding plan required.
- Type-safe domain models in `services/pkg/domain/` shared across services.
