output "key_ring_id" {
  description = "ID of the platform KMS keyring."
  value       = google_kms_key_ring.platform.id
}

output "bigquery_key_id" {
  description = "BigQuery CMEK ID."
  value       = google_kms_crypto_key.bigquery.id
}

output "gcs_key_id" {
  description = "GCS CMEK ID."
  value       = google_kms_crypto_key.gcs.id
}

output "pubsub_key_id" {
  description = "Pub/Sub CMEK ID."
  value       = google_kms_crypto_key.pubsub.id
}

output "alloydb_key_id" {
  description = "AlloyDB CMEK ID."
  value       = google_kms_crypto_key.alloydb.id
}

output "vertex_key_id" {
  description = "Vertex AI CMEK ID."
  value       = google_kms_crypto_key.vertex.id
}

output "fraud_runtime_sa" {
  description = "Service account email for fraud-service runtime."
  value       = google_service_account.fraud_runtime.email
}

output "risk_runtime_sa" {
  description = "Service account email for risk-service runtime."
  value       = google_service_account.risk_runtime.email
}

output "chat_runtime_sa" {
  description = "Service account email for chat-bff runtime."
  value       = google_service_account.chat_runtime.email
}

output "doc_runtime_sa" {
  description = "Service account email for document-service runtime."
  value       = google_service_account.doc_runtime.email
}

output "pipeline_runner_sa" {
  description = "Service account email for Vertex AI pipeline runner."
  value       = google_service_account.pipeline_runner.email
}

output "armor_policy_self_link" {
  description = "Self-link of the Cloud Armor security policy."
  value       = google_compute_security_policy.platform.self_link
}
