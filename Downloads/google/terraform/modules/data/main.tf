###############################################################################
# BigQuery dataset.
###############################################################################

resource "google_bigquery_dataset" "platform" {
  dataset_id  = "finguard_${var.environment}"
  description = "FinGuard AI data warehouse (${var.environment})."
  location    = var.region

  default_encryption_configuration {
    kms_key_name = var.bigquery_kms_key_id
  }

  labels = {
    env        = var.environment
    managed_by = "terraform"
  }
}

###############################################################################
# BigQuery tables — schemas are loaded from schemas/*.sql files which
# contain `CREATE TABLE` statements. We extract the DDL with file()
# and apply it via the BigQuery Jobs API.
###############################################################################

locals {
  schemas_dir = "${path.module}/schemas"
  tables = {
    transactions = "transactions.sql"
    risk_scores  = "risk_scores.sql"
    documents    = "documents.sql"
    alerts       = "alerts.sql"
    audit_log    = "audit_log.sql"
  }
}

resource "google_bigquery_table" "tables" {
  for_each = local.tables

  dataset_id = google_bigquery_dataset.platform.dataset_id
  table_id   = each.key

  # Read the DDL from disk. The file must contain a single
  # `CREATE TABLE ...` statement for the corresponding table_id.
  deletion_protection = var.environment == "prod" ? true : false

  # Schema is parsed from the DDL at apply-time by Terraform; we keep the
  # full DDL in a per-table sql file under schemas/ for reference.
  schema = file("${local.schemas_dir}/${each.key}.json")
}

# Note: DDL files in `schemas/` are applied out-of-band by the data team's
# `bq` workflow. We keep them in this module so the versioned schema and the
# Terraform-declared resource attributes stay in lock-step, but we do NOT
# run them inside Terraform — that would couple terraform apply to a developer
# gcloud auth and make CI risky.

###############################################################################
# Pub/Sub topics.
###############################################################################

resource "google_pubsub_topic" "txn_events" {
  name    = "txn-events"
  labels  = { env = var.environment }

  message_retention_duration = "604800s"

  kms_key_name = var.pubsub_kms_key_id
}

resource "google_pubsub_topic" "txn_scoring" {
  name    = "txn-scoring"
  labels  = { env = var.environment }

  message_retention_duration = "604800s"

  kms_key_name = var.pubsub_kms_key_id
}

resource "google_pubsub_topic" "doc_intake" {
  name    = "doc-intake"
  labels  = { env = var.environment }

  message_retention_duration = "604800s"

  kms_key_name = var.pubsub_kms_key_id
}

resource "google_pubsub_topic" "alert_fanout" {
  name    = "alert-fanout"
  labels  = { env = var.environment }

  message_retention_duration = "604800s"

  kms_key_name = var.pubsub_kms_key_id
}

# Dead-letter topics (one per primary topic).
resource "google_pubsub_topic" "txn_events_dlq" {
  name   = "txn-events-dlq"
  labels = { env = var.environment }
  kms_key_name = var.pubsub_kms_key_id
}

resource "google_pubsub_topic" "doc_intake_dlq" {
  name   = "doc-intake-dlq"
  labels = { env = var.environment }
  kms_key_name = var.pubsub_kms_key_id
}

###############################################################################
# GCS buckets.
###############################################################################

resource "google_storage_bucket" "documents" {
  name     = "finguard-${var.environment}-documents"
  location = var.region

  uniform_bucket_level_access = true
  force_destroy               = var.environment != "prod"

  encryption {
    default_kms_key_name = var.gcs_kms_key_id
  }

  versioning {
    enabled = true
  }

  lifecycle_rule {
    action {
      type = "SetStorageClass"
      storage_class = "ARCHIVE"
    }
    condition {
      age = 90
    }
  }

  labels = {
    env = var.environment
  }
}

resource "google_storage_bucket" "ml_artifacts" {
  name     = "finguard-${var.environment}-ml-artifacts"
  location = var.region

  uniform_bucket_level_access = true
  force_destroy               = var.environment != "prod"

  encryption {
    default_kms_key_name = var.gcs_kms_key_id
  }

  versioning {
    enabled = true
  }

  labels = {
    env = var.environment
  }
}

resource "google_storage_bucket" "dataflow_staging" {
  name     = "finguard-${var.environment}-dataflow-staging"
  location = var.region

  uniform_bucket_level_access = true
  force_destroy               = true # ephemeral

  labels = {
    env = var.environment
  }
}
