# Terraform Bootstrap

Some resources must be created **before** Terraform can manage the rest of the
infrastructure (chicken-and-egg). These are documented here as manual steps
run **once** per Google Cloud organization.

## Prerequisites

- A Google Cloud organization (or a sandbox project for early development)
- Owner / Org Admin role on the bootstrap account
- `gcloud` CLI installed and authenticated
- A unique GCS bucket name for Terraform state (e.g. `finguard-tfstate-<org-id>`)

## Step 1 — Create the state bucket

```bash
PROJECT_ID="finguard-bootstrap"
REGION="us-central1"

gcloud projects create $PROJECT_ID --name="FinGuard Bootstrap"
gcloud config set project $PROJECT_ID

# Enable required APIs
gcloud services enable \
  cloudresourcemanager.googleapis.com \
  storage.googleapis.com \
  iam.googleapis.com

# Create state bucket (versioned, with uniform bucket-level access)
gsutil mb -l $REGION -b on gs://finguard-tfstate-$PROJECT_ID
gsutil versioning set on gs://finguard-tfstate-$PROJECT_ID
```

## Step 2 — Enable APIs in target project

The same APIs needed by the modules must be enabled in the project Terraform
will manage. The list is owned by `envs/<env>/main.tf`; the bootstrap project
only needs the state-management APIs above.

## Step 3 — Configure remote backend

For each environment, copy `envs/<env>/backend.tf.example` to `backend.tf` and
fill in the bucket name:

```hcl
terraform {
  backend "gcs" {
    bucket = "finguard-tfstate-<org-id>"
    prefix = "envs/dev"
  }
}
```

## Step 4 — Grant Terraform Service Account

Create a service account used by Cloud Build to run Terraform, with the
following roles on the target project:

- `roles/owner` (or narrower roles per-module — TBD)
- `roles/iam.serviceAccountTokenCreator` (for impersonation)

```bash
gcloud iam service-accounts create terraform-runner \
  --display-name="Terraform Runner"

gcloud projects add-iam-policy-binding $TARGET_PROJECT \
  --member="serviceAccount:terraform-runner@$TARGET_PROJECT.iam.gserviceaccount.com" \
  --role="roles/owner"
```

## Step 5 — First apply

```bash
cd envs/dev
terraform init
terraform plan -out tfplan
terraform apply tfplan
```

After the first successful apply, all subsequent changes flow through
Cloud Build (`deploy/cloudbuild/terraform-plan.yaml`,
`deploy/cloudbuild/terraform-apply.yaml`).

## What CAN'T be terraformed

- **Organization creation** — must be done via Google Workspace / Cloud Identity.
- **Initial billing account attachment** — must be done in the console.
- **VPC-SC perimeter containing the bootstrap project** — chicken-and-egg; once
  the perimeter exists, exclude the bootstrap project so state can still be read.

These are intentional boundaries; see `docs/security.md` for details.
