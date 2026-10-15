variable "project_id" {
  description = "GCP project ID for the dev environment."
  type        = string
}

variable "region" {
  description = "GCP region."
  type        = string
  default     = "us-central1"
}

variable "access_policy_id" {
  description = "Numeric ID of the org-level Access Context Manager policy."
  type        = string
}

variable "alert_email" {
  description = "Email for monitoring/drift alerts."
  type        = string
  default     = "dev-alerts@example.com"
}
