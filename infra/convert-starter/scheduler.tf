resource "google_cloud_scheduler_job" "convert_starter_schedule" {
  name             = "convert-starter-schedule-job"
  description      = "Trigger convert-starter function every 10 minutes"
  schedule         = "*/10 * * * *"
  time_zone        = "Asia/Tokyo"
  pubsub_target {
    topic_name = google_pubsub_topic.convert_starter_schedule_topic.id
    data       = base64encode("{\"message\":\"trigger\"}")
  }
}
