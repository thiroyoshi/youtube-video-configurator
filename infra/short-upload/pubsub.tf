resource "google_pubsub_topic" "short_upload_schedule_topic" {
  name = "short-upload-schedule-topic"
}

resource "google_pubsub_topic_iam_member" "pubsub_publisher" {
  project = var.project_id
  topic   = google_pubsub_topic.short_upload_schedule_topic.name
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:service-${var.project_number}@gcp-sa-cloudscheduler.iam.gserviceaccount.com"
}

resource "google_pubsub_subscription" "short_upload_subscription" {
  name  = "short-upload-subscription"
  topic = google_pubsub_topic.short_upload_schedule_topic.name

  ack_deadline_seconds = 20

  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }

  expiration_policy {
    ttl = ""
  }
}
