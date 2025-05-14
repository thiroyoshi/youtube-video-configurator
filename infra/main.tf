resource "google_project_service" "cloud_scheduler" {
  project = var.project_id
  service = "cloudscheduler.googleapis.com"
}

resource "time_sleep" "wait_for_scheduler_api" {
  depends_on = [google_project_service.cloud_scheduler]
  create_duration = "30s"
}

module "convert-starter" {
  source        = "./convert-starter"
  project_id    = var.project_id
  region        = var.region
  source_bucket = var.source_bucket

  depends_on = [time_sleep.wait_for_scheduler_api]
}
