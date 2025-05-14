output "function_name" {
  value = google_cloudfunctions2_function.video_converter.name
}

output "service_account_email" {
  value = google_service_account.video_converter_sa.email
}
