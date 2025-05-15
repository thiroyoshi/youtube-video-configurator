resource "google_cloud_scheduler_job" "blog_post_diary_schedule" {
  name        = "blog-post-diary-schedule-job"
  description = "Trigger blog-post-diary function every day at 0:00 JST"
  schedule    = "0 0 * * *"
  time_zone   = "Asia/Tokyo"
  region      = var.region
  project     = var.project_id

  retry_config {
    retry_count = 1
  }

  pubsub_target {
    topic_name = google_pubsub_topic.blog_post_diary_schedule_topic.id
    data       = base64encode("{\"message\":\"trigger\"}")
  }
}

resource "google_service_account" "scheduler_sa" {
  account_id   = "blog-post-diary-scheduler-sa"
  display_name = "Service Account for Blog Post Diary Cloud Scheduler"
}

resource "google_project_iam_member" "scheduler_pubsub_publisher" {
  project = var.project_id
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:${google_service_account.scheduler_sa.email}"
}
