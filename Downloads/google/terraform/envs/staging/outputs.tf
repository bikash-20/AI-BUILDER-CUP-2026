output "gateway_hostname" {
  description = "Public hostname of the API Gateway for the staging env."
  value       = module.apigateway.gateway_default_hostname
}

output "fraud_service_url" {
  description = "Cloud Run URL of the fraud service."
  value       = module.cloud_run.service_urls["fraud-service"]
}

output "vpc_id" {
  description = "Staging VPC ID."
  value       = module.network.vpc_id
}

output "bq_dataset" {
  description = "BigQuery dataset ID for the platform."
  value       = module.data.dataset_id
}
