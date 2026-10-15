output "gateway_hostname" {
  description = "Public hostname of the API Gateway for the prod env."
  value       = module.apigateway.gateway_default_hostname
  sensitive   = true
}

output "fraud_service_url" {
  description = "Cloud Run URL of the fraud service."
  value       = module.cloud_run.service_urls["fraud-service"]
  sensitive   = true
}

output "vpc_id" {
  description = "Prod VPC ID."
  value       = module.network.vpc_id
}

output "bq_dataset" {
  description = "BigQuery dataset ID for the platform."
  value       = module.data.dataset_id
}

output "alloydb_cluster" {
  description = "AlloyDB cluster name."
  value       = module.alloydb.cluster_id
}
