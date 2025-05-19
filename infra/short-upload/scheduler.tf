resource "google_cloud_scheduler_job" "short_upload_scheduler" {
  name             = "short-upload-scheduler"
  description      = "Triggers the short-upload Cloud Function every hour"
  schedule         = "0 * * * *"  # 毎時0分に実行
  time_zone        = "Asia/Tokyo"
  attempt_deadline = "320s"

  pubsub_target {
    topic_name = google_pubsub_topic.short_upload_schedule_topic.id
    data       = base64encode("{}")
  }
}
