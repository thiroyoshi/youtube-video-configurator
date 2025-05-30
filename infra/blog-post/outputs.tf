output "function_name" {
  value = google_cloudfunctions2_function.blog_post.name
}

output "function_url" {
  value       = google_cloudfunctions2_function.blog_post.service_config[0].uri
  description = "The HTTPS endpoint of the deployed Cloud Function."
}

output "service_account_email" {
  value = google_service_account.function_sa.email
}
