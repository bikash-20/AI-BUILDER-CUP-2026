variable "project_id" {
  description = "GCP project ID."
  type        = string
}

variable "environment" {
  description = "Environment name (dev / staging / prod)."
  type        = string
}

variable "region" {
  description = "GCP region for the API Gateway."
  type        = string
  default     = "us-central1"
}

variable "gateway_runtime_sa" {
  description = "Service account used by API Gateway to invoke Cloud Run services."
  type        = string
}

variable "fraud_service_url" {
  description = "Cloud Run URL of the fraud-service."
  type        = string
}

variable "risk_service_url" {
  description = "Cloud Run URL of the risk-service."
  type        = string
}

variable "chat_bff_url" {
  description = "Cloud Run URL of the chat-bff."
  type        = string
}

variable "doc_service_url" {
  description = "Cloud Run URL of the document-service."
  type        = string
}
