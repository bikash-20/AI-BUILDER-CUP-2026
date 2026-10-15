variable "project_id" {
  description = "GCP project ID."
  type        = string
}

variable "environment" {
  description = "Environment name (dev / staging / prod)."
  type        = string
}

variable "primary_region" {
  description = "Primary region for KMS keyring location."
  type        = string
  default     = "us-central1"
}

variable "access_policy_id" {
  description = "Numeric ID of the org-level Access Context Manager policy."
  type        = string
}
