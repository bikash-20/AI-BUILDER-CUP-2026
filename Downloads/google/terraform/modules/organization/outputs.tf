output "finguard_folder_id" {
  description = "ID of the top-level FinGuard AI folder."
  value       = google_folder.finguard.id
}

output "dev_folder_id" {
  description = "ID of the Dev environment folder."
  value       = google_folder.dev.id
}

output "staging_folder_id" {
  description = "ID of the Staging environment folder."
  value       = google_folder.staging.id
}

output "prod_folder_id" {
  description = "ID of the Prod environment folder."
  value       = google_folder.prod.id
}
