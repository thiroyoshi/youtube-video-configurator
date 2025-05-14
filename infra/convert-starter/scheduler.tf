resource "google_cloud_scheduler_job" "convert_starter_schedule" {
  name             = "convert-starter-schedule-job"
  description      = "Trigger convert-starter function every 10 minutes"
  schedule         = "*/10 * * * *"
  time_zone        = "Asia/Tokyo"
  attempt_deadline = "320s"
  region           = var.region
  project          = var.project_id

  retry_config {
    retry_count = 1
  }

  pubsub_target {
    topic_name = google_pubsub_topic.convert_starter_schedule_topic.id
    data       = base64encode("{\"message\":\"trigger\"}")
  }
}

# Cloud Scheduler用のサービスアカウント
resource "google_service_account" "scheduler_sa" {
  account_id   = "convert-starter-scheduler-sa"
  display_name = "Service Account for Convert Starter Cloud Scheduler"
}

# Cloud SchedulerのサービスアカウントにPub/Subへの発行権限を付与
resource "google_project_iam_member" "scheduler_pubsub_publisher" {
  project = var.project_id
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:${google_service_account.scheduler_sa.email}"
}
