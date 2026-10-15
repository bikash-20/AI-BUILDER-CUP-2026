output "cluster_id" {
  description = "AlloyDB cluster name."
  value       = google_alloydb_cluster.platform.name
}

output "primary_instance_id" {
  description = "AlloyDB primary instance name."
  value       = google_alloydb_instance.primary.name
}

output "primary_ip" {
  description = "Private IP of the primary instance (via service networking)."
  value       = google_alloydb_instance.primary.ip_address
  sensitive   = true
}
