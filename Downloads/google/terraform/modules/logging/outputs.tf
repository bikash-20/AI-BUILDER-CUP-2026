output "logs_dataset_id" {
  description = "BigQuery dataset ID for logs."
  value       = google_bigquery_dataset.logs.dataset_id
}

output "vpc_flow_topic" {
  description = "Pub/Sub topic name for VPC flow logs."
  value       = google_pubsub_topic.vpc_flow.name
}

output "fraud_decisions_metric" {
  description = "Log-based metric name for fraud decisions."
  value       = google_logging_metric.fraud_decisions.name
}
