output "secret_ids" {
  description = "Map of created secret IDs"
  value       = { for k, v in google_secret_manager_secret.secrets : k => v.secret_id }
}
