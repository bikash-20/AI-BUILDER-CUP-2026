output "vpc_id" {
  description = "ID of the environment VPC."
  value       = google_compute_network.vpc.id
}

output "vpc_self_link" {
  description = "Self-link of the environment VPC."
  value       = google_compute_network.vpc.self_link
}

output "primary_subnet_id" {
  description = "ID of the primary subnet."
  value       = google_compute_subnetwork.primary.id
}

output "dr_subnet_id" {
  description = "ID of the DR subnet."
  value       = google_compute_subnetwork.dr.id
}

output "private_services_range" {
  description = "Name of the reserved private services range."
  value       = google_compute_global_address.private_services.name
}
