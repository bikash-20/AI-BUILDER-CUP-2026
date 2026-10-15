output "featurestore_id" {
  description = "Vertex AI Featurestore ID."
  value       = google_vertex_ai_featurestore.platform.id
}

output "fraud_endpoint_id" {
  description = "Vertex AI Endpoint for the fraud model."
  value       = google_vertex_ai_endpoint.fraud.id
}

output "fraud_endpoint_resource_name" {
  description = "Full resource name of the fraud endpoint (projects/.../endpoints/...)."
  value       = google_vertex_ai_endpoint.fraud.name
}

output "customer_entity_id" {
  description = "Featurestore entity type for customers."
  value       = google_vertex_ai_featurestore_entitytype.customer.id
}

output "transaction_entity_id" {
  description = "Featurestore entity type for transactions."
  value       = google_vertex_ai_featurestore_entitytype.transaction.id
}

output "device_entity_id" {
  description = "Featurestore entity type for devices."
  value       = google_vertex_ai_featurestore_entitytype.device.id
}
