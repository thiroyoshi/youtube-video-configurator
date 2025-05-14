resource "google_project_service" "serviceusage" {
  project = var.project_id
  service = "serviceusage.googleapis.com"
}

resource "google_project_service" "cloud_scheduler" {
  project = var.project_id
  service = "cloudscheduler.googleapis.com"
  depends_on = [google_project_service.serviceusage]
}

resource "google_project_service" "pubsub" {
  project = var.project_id
  service = "pubsub.googleapis.com"
  depends_on = [google_project_service.serviceusage]
}

resource "google_service_account" "cloudbuild_sa" {
  account_id   = "cloudbuild-trigger-sa"
  display_name = "Service Account for Cloud Build Trigger"
}

locals {
  cloudbuild_roles = [
    "roles/pubsub.admin",
    "roles/cloudfunctions.admin",
    "roles/cloudscheduler.admin",
    "roles/resourcemanager.projectIamAdmin"
  ]
}

resource "google_project_iam_member" "cloudbuild_sa_roles" {
  for_each = toset(local.cloudbuild_roles)
  project = var.project_id
  role    = each.value
  member  = "serviceAccount:${google_service_account.cloudbuild_sa.email}"
}

resource "time_sleep" "wait_for_scheduler_api" {
  depends_on = [google_project_service.cloud_scheduler]
  create_duration = "30s"
}

# module "convert-starter_deploy_trigger" {
#   source         = "./cloudbuild_trigger"
#   trigger_name   = "convert-starter-deploy-trigger"
#   function_name  = "convert-starter"
#   cloudbuild_sa_email = google_service_account.cloudbuild_sa.email
# }

# module "video-converter_deploy_trigger" {
#   source         = "./cloudbuild_trigger"
#   trigger_name   = "video-converter-deploy-trigger"
#   function_name  = "video-converter"
#   cloudbuild_sa_email = google_service_account.cloudbuild_sa.email
# }

module "convert-starter" {
  source        = "./convert-starter"
  project_id    = var.project_id
  project_number = var.project_number
  region        = var.region
  source_bucket = var.source_bucket
  depends_on = [time_sleep.wait_for_scheduler_api, google_project_service.pubsub]
}

module "video-converter" {
  source        = "./video-converter"
  project_id    = var.project_id
  region        = var.region
  source_bucket = var.source_bucket
  convert_starter_service_account_email = module.convert-starter.service_account_email
}
