variable "project_id" {
  description = "GCP project ID."
  type        = string
}

variable "environment" {
  description = "Environment name (dev / staging / prod)."
  type        = string
}

variable "region" {
  description = "GCP region for resources."
  type        = string
  default     = "us-central1"
}

variable "bigquery_kms_key_id" {
  description = "BigQuery CMEK key ID from the security module."
  type        = string
}

variable "gcs_kms_key_id" {
  description = "GCS CMEK key ID from the security module."
  type        = string
}

variable "pubsub_kms_key_id" {
  description = "Pub/Sub CMEK key ID from the security module."
  type        = string
}
