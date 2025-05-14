resource "google_service_account" "function_sa" {
  account_id   = "convert-starter-fn-sa"
  display_name = "Service Account for Convert Starter Cloud Function"
}

resource "google_project_iam_member" "cloudbuild_pubsub_admin" {
  project = var.project_id
  role    = "roles/pubsub.admin"
  member  = "serviceAccount:${var.project_number}@cloudbuild.gserviceaccount.com"
}
