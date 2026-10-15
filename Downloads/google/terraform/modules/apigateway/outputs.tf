output "gateway_id" {
  description = "API Gateway resource name."
  value       = google_api_gateway_gateway.platform.id
}

output "gateway_default_hostname" {
  description = "Default hostname of the API Gateway."
  value       = google_api_gateway_gateway.platform.default_hostname
}
