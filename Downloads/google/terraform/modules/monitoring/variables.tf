variable "project_id" {
  description = "GCP project ID."
  type        = string
}

variable "environment" {
  description = "Environment name (dev / staging / prod)."
  type        = string
}

variable "alert_email" {
  description = "Email address for monitoring notifications."
  type        = string
  default     = "platform-alerts@example.com"
}

variable "alert_pubsub_topic" {
  description = "Pub/Sub topic for PagerDuty bridge."
  type        = string
  default     = "alerts-pagerduty"
}
