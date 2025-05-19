output "function_uri" {
  value       = google_cloudfunctions2_function.short_upload.service_config[0].uri
  description = "The URI of the deployed function"
}

output "service_account_email" {
  value       = google_service_account.function_sa.email
  description = "The email of the service account"
}
