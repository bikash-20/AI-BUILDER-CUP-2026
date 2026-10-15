output "dataset_id" {
  description = "BigQuery dataset ID for the platform."
  value       = google_bigquery_dataset.platform.dataset_id
}

output "txn_events_topic" {
  description = "Pub/Sub topic for transaction events."
  value       = google_pubsub_topic.txn_events.name
}

output "txn_scoring_topic" {
  description = "Pub/Sub topic for transaction scoring requests."
  value       = google_pubsub_topic.txn_scoring.name
}

output "doc_intake_topic" {
  description = "Pub/Sub topic for document intake events."
  value       = google_pubsub_topic.doc_intake.name
}

output "alert_fanout_topic" {
  description = "Pub/Sub topic for alert fanout to UI."
  value       = google_pubsub_topic.alert_fanout.name
}

output "documents_bucket" {
  description = "GCS bucket name for uploaded documents."
  value       = google_storage_bucket.documents.name
}

output "ml_artifacts_bucket" {
  description = "GCS bucket name for ML training artifacts."
  value       = google_storage_bucket.ml_artifacts.name
}
