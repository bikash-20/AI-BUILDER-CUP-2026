output "service_urls" {
  description = "Map of service name → Cloud Run URL."
  value = {
    for k, s in google_cloud_run_v2_service.services :
    k => s.uri
  }
}

output "service_names" {
  description = "Map of service name → Cloud Run resource name."
  value = {
    for k, s in google_cloud_run_v2_service.services :
    k => s.name
  }
}

output "vpc_connector_id" {
  description = "VPC connector ID."
  value       = google_vpc_access_connector.connector.id
}
