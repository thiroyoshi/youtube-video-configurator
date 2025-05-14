resource "google_service_account" "function_sa" {
  account_id   = "convert-starter-fn-sa"
  display_name = "Service Account for Convert Starter Cloud Function"
}
