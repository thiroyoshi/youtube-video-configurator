resource "google_service_account" "video_converter_sa" {
  account_id   = "video-converter-fn-sa"
  display_name = "Service Account for Video Converter Cloud Function"
}

locals {
  cloudbuild_roles = [
    "roles/pubsub.admin",
    "roles/cloudfunctions.admin",
    "roles/cloudscheduler.admin",
    "roles/resourcemanager.projectIamAdmin"
  ]
}

resource "google_project_iam_member" "cloudbuild_roles" {
  for_each = toset(local.cloudbuild_roles)
  project = var.project_id
  role    = each.value
  member  = "serviceAccount:${var.project_number}@cloudbuild.gserviceaccount.com"
}
