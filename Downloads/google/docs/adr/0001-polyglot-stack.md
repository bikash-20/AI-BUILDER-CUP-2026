# ADR-0001 — Polyglot Stack (Go + Python + TypeScript)

- **Status:** Accepted
- **Date:** 2026-10-15
- **Deciders:** Architecture Guild

## Context

FinGuard AI requires both low-latency transaction scoring (< 500 ms P95) and rich ML/AI workloads (XGBoost, autoencoders, Gemini, Document AI, Beam). A single-language stack would either be slow in the hot path or lack the ML/AI ecosystem.

## Decision

Adopt a **polyglot stack** aligned to workload:

| Layer | Language | Why |
|---|---|---|
| Frontend | TypeScript + Next.js | Type safety, SSR, RSC, large ecosystem |
| Cloud Run services (latency-critical) | Go | Sub-millisecond overhead, strong concurrency, single binary deploy |
| ML / AI services (Gemini, DocAI, eval) | Python | First-class Google AI SDKs, Beam, scikit-learn, Vertex AI SDK |
| Data pipelines | Python (Beam) | Apache Beam native, Dataflow runner |
| IaC | Terraform | GCP provider maturity, state management |

## Consequences

**Positive**
- Right tool for each job
- Independent scaling and release cadence
- Hireable skills in every layer

**Negative**
- More language toolchains to maintain
- Cross-service observability requires care (correlation IDs, OTel)
- Repos need to be aware of multiple CI pipelines

**Mitigations**
- Shared libraries (`services/pkg/`, `ml/libs/finguard_common/`)
- Standardized OTel setup in each runtime
- One Makefile at the root, per-folder Make targets
