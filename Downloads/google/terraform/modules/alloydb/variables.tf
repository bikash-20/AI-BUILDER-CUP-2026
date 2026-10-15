variable "project_id" {
  description = "GCP project ID."
  type        = string
}

variable "environment" {
  description = "Environment name (dev / staging / prod)."
  type        = string
}

variable "region" {
  description = "GCP region for the cluster."
  type        = string
  default     = "us-central1"
}

variable "network_id" {
  description = "Self-link of the VPC the cluster attaches to."
  type        = string
}

variable "alloydb_kms_key_id" {
  description = "AlloyDB CMEK key ID."
  type        = string
}
