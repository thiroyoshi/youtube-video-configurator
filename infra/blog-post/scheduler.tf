resource "google_cloud_scheduler_job" "blog_post_schedule" {
  name        = "blog-post-schedule-job"
  description = "Trigger blog-post function on Mon, Wed, Fri at 6:00 JST"
  schedule    = "0 6 * * 1,3,5"
  time_zone   = "Asia/Tokyo"
  region      = var.region
  project     = var.project_id

  retry_config {
    retry_count = 1
  }

  pubsub_target {
    topic_name = google_pubsub_topic.blog_post_schedule_topic.id
    data       = base64encode("{\"message\":\"trigger\"}")
  }
}

resource "google_service_account" "scheduler_sa" {
  account_id   = "blog-post-scheduler-sa"
  display_name = "Service Account for Blog Post Cloud Scheduler"
}

resource "google_project_iam_member" "scheduler_pubsub_publisher" {
  project = var.project_id
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:${google_service_account.scheduler_sa.email}"
}
