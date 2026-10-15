output "email_channel_id" {
  description = "ID of the email notification channel."
  value       = google_monitoring_notification_channel.email.id
}

output "pubsub_channel_id" {
  description = "ID of the Pub/Sub notification channel."
  value       = google_monitoring_notification_channel.pubsub.id
}

output "fraud_latency_slo_id" {
  description = "SLO ID for fraud latency."
  value       = google_monitoring_slo.fraud_latency.id
}

output "fraud_availability_slo_id" {
  description = "SLO ID for fraud availability."
  value       = google_monitoring_slo.fraud_availability.id
}
