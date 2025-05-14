resource "google_pubsub_topic" "convert_starter_schedule_topic" {
  name = "convert-starter-schedule-topic"
}

resource "google_project_iam_member" "pubsub_invoker" {
  project = var.project_id
  role    = "roles/cloudfunctions.invoker"
  member  = "serviceAccount:service-${var.project_number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}

# Cloud FunctionsのサービスアカウントにPub/Subのsubscriber権限を付与
resource "google_pubsub_topic_iam_member" "function_subscriber" {
  topic  = google_pubsub_topic.convert_starter_schedule_topic.name
  role   = "roles/pubsub.subscriber"
  member = "serviceAccount:${google_service_account.function_sa.email}"
}

# Cloud Pub/SubサービスアカウントにPub/Subのpublisher権限を付与
resource "google_pubsub_topic_iam_member" "pubsub_publisher" {
  topic  = google_pubsub_topic.convert_starter_schedule_topic.name
  role   = "roles/pubsub.publisher"
  member = "serviceAccount:service-${var.project_number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}