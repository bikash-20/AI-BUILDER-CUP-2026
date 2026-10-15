variable "project_id" {
  description = "GCP project ID."
  type        = string
}

variable "environment" {
  description = "Environment name (dev / staging / prod)."
  type        = string
}

variable "region" {
  description = "GCP region for Vertex AI resources."
  type        = string
  default     = "us-central1"
}

variable "vertex_kms_key_id" {
  description = "Vertex AI CMEK key ID from the security module."
  type        = string
}

variable "alert_email" {
  description = "Email for drift alerts."
  type        = string
  default     = "ml-alerts@example.com"
}
