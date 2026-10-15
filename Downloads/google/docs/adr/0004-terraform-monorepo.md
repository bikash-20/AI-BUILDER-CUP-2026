# ADR-0004 — Terraform Monorepo with Module Hierarchy

- **Status:** Accepted
- **Date:** 2026-10-15

## Context

FinGuard's infrastructure spans organization, network, security, data, Vertex AI, Cloud Run, monitoring, and API gateway. Splitting Terraform across multiple repos creates drift; one giant root module becomes unmaintainable.

## Decision

Adopt a **Terraform monorepo** with the following layout:

```
terraform/
├── bootstrap/         # manual one-time setup
├── modules/           # reusable, environment-agnostic
│   ├── organization/  # folders, org policies
│   ├── network/       # VPC, subnets, NAT, firewall
│   ├── security/      # KMS, VPC-SC, Cloud Armor
│   ├── logging/       # log sinks, BQ dataset
│   ├── data/          # BigQuery, Pub/Sub, GCS
│   ├── vertex_ai/     # Featurestore, Endpoint, Monitor
│   ├── cloud_run/     # Services + IAM
│   ├── alloydb/       # instance + DBs
│   ├── apigateway/    # API GW + OpenAPI
│   └── monitoring/    # SLOs, alerts, dashboards
└── envs/              # environment-specific compositions
    ├── dev/
    ├── staging/
    └── prod/
```

Each module is independently versioned (via tags) and consumed by envs.

## Consequences

**Positive**
- One PR can update infra + service in lockstep
- Modules are reusable; promote via tags
- Easy to test with `terraform validate` in CI

**Negative**
- Larger blast radius — partial failures more likely
- Need discipline to keep modules reusable
- State locking required (GCS backend)

**Mitigations**
- Strict module boundaries (no cross-module references)
- Mandatory code review for `envs/prod/`
- Per-env state files
- Drift detection job in Cloud Scheduler
