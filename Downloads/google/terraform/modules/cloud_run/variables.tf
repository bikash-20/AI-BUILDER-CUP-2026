variable "project_id" {
  description = "GCP project ID."
  type        = string
}

variable "environment" {
  description = "Environment name (dev / staging / prod)."
  type        = string
}

variable "region" {
  description = "GCP region for Cloud Run."
  type        = string
  default     = "us-central1"
}

variable "network_id" {
  description = "VPC self-link for the VPC connector."
  type        = string
}

# Service-specific inputs

variable "fraud_image" {
  description = "Initial fraud-service container image (will be replaced by Cloud Deploy)."
  type        = string
  default     = "us-docker.pkg.dev/cloudrun/hello:latest"
}

variable "risk_image" {
  description = "Initial risk-service container image."
  type        = string
  default     = "us-docker.pkg.dev/cloudrun/hello:latest"
}

variable "chat_bff_image" {
  description = "Initial chat-bff container image."
  type        = string
  default     = "us-docker.pkg.dev/cloudrun/hello:latest"
}

variable "doc_image" {
  description = "Initial document-service container image."
  type        = string
  default     = "us-docker.pkg.dev/cloudrun/hello:latest"
}

variable "fraud_runtime_sa" {
  description = "Service account email for fraud-service runtime."
  type        = string
}

variable "risk_runtime_sa" {
  description = "Service account email for risk-service runtime."
  type        = string
}

variable "chat_runtime_sa" {
  description = "Service account email for chat-bff runtime."
  type        = string
}

variable "doc_runtime_sa" {
  description = "Service account email for document-service runtime."
  type        = string
}

variable "gateway_runtime_sa" {
  description = "Service account for the API Gateway runtime."
  type        = string
}

variable "vertex_fraud_endpoint" {
  description = "Vertex AI endpoint resource name for fraud scoring."
  type        = string
}

variable "feature_store_id" {
  description = "Vertex AI Featurestore ID."
  type        = string
}

variable "gemini_service_url" {
  description = "Internal URL of the Gemini service."
  type        = string
  default     = ""
}

variable "docai_processor_id" {
  description = "Document AI processor resource name."
  type        = string
  default     = ""
}
