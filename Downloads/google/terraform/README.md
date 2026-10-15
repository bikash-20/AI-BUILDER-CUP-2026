# FinGuard AI — Infrastructure as Code (Terraform)

This directory contains all Google Cloud infrastructure defined declaratively with Terraform.
The layout follows a **module-per-concern** pattern composed per environment
(`dev`, `staging`, `prod`) under `envs/`.

## Layout

```
terraform/
├── bootstrap/                    # Manual bootstrap (state bucket, APIs)
├── modules/                      # Reusable, single-purpose modules
│   ├── organization/             # Folders, org policies, billing
│   ├── network/                  # VPC, subnets, NAT, Private Service Access
│   ├── security/                 # KMS, VPC-SC, Cloud Armor, IAM
│   ├── logging/                  # Log sinks, BQ dataset, retention
│   ├── data/                     # BigQuery, Pub/Sub, GCS, schemas/
│   ├── vertex_ai/                # Feature store, endpoint, registry, monitor
│   ├── cloud_run/                # Cloud Run services + IAM
│   ├── monitoring/               # SLOs, alerts, dashboards/
│   ├── apigateway/               # API Gateway + openapi.yaml
│   └── alloydb/                  # AlloyDB cluster + instance
└── envs/                         # One composition per environment
    ├── dev/
    ├── staging/
    └── prod/
```

## Usage

```bash
# 1) One-time bootstrap (creates state bucket, enables APIs)
cd bootstrap && cat README.md

# 2) Select an environment
cd ../envs/dev

# 3) Authenticate and initialize
gcloud auth application-default login
terraform init

# 4) Plan and review
terraform plan -out tfplan

# 5) Apply
terraform apply tfplan
```

## Conventions

- **Pin providers** with `versions.tf` in every module (`google ~> 5.30`, `google-beta ~> 5.30`).
- **Pin module sources** in `envs/` to the local path (`source = "../../modules/network"`).
- **Naming** follows `<project>-<env>-<region>-<resource>` (e.g. `finguard-dev-us-central1-fraud-svc`).
- **Labels** are propagated to all resources via a `labels` map (`env`, `service`, `team`, `managed_by`).
- **Remote state** uses GCS — backend is configured per env in `envs/<env>/backend.tf`.
- **Sensitive outputs** are marked `sensitive = true` and never printed in CI logs.

## CI/CD

`deploy/cloudbuild/terraform-plan.yaml` runs `terraform plan` on PRs.
`deploy/cloudbuild/terraform-apply.yaml` runs `terraform apply` on merges to `main`
after a manual approval gate (Cloud Build approval).

See `docs/security.md` for VPC-SC perimeter setup and CMEK keyring topology.
