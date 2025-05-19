resource "google_cloudfunctions2_function" "short_upload" {
  name     = "short-upload"
  location = var.region
  build_config {
    runtime     = "go123"
    entry_point = "shortUpload"
    source {
      storage_source {
        bucket = var.source_bucket
        object = "short-upload_${var.short_sha}.zip"
      }
    }
  }
  service_config {
    service_account_email = google_service_account.function_sa.email
    environment_variables = {
      GOOGLE_CLOUD_PROJECT  = var.project_id
      X_API_KEY             = "vR8oo1pAQFgeYKlfxIPSrgRq6"
      X_API_SECRET_KEY      = "fyS3Nm8tEsSQOKK9Ez77TQn7Fi2A3HSO7ZdkDAArshXCSxNXT0"
      X_ACCESS_TOKEN        = "1449548285354516482-BxphqsVkM9LQUjHzIVpHnJ2DqcGQTw"
      X_ACCESS_TOKEN_SECRET = "1fj79P9ttUavCvjH7iZGVITuTgbqx5VqgrEznLPJTsVvU"
      X_USER_ID             = "1449548285354516482"
      YOUTUBE_CLIENT_ID     = "589350762095-2rpqdftrm5m5s0ibhg6m1kb0f46q058r.apps.googleusercontent.com"
      YOUTUBE_CLIENT_SECRET = "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"
      YOUTUBE_REFRESH_TOKEN = "1//0eZ6zn_HG54e-CgYIARAAGA4SNwF-L9IraHLGPq_CNydexr-Sjj0SczlZZF0M3r6A5Sp2O8Eo_1tnR7mUUeFPpRIJ2v87_8QeHEI"
      PLAYLIST_SHORT        = "PLTSYDCu3sM9LEQ27HYpSlCMrxHyquc-_O"
      FORTNITE_SEASON       = "C6S3"
    }
    min_instance_count             = 0
    max_instance_count             = 1
    available_memory               = "256M"
    timeout_seconds                = 540
    ingress_settings               = "ALLOW_ALL"
    all_traffic_on_latest_revision = true
  }

  event_trigger {
    trigger_region = var.region
    event_type     = "google.cloud.pubsub.topic.v1.messagePublished"
    pubsub_topic   = google_pubsub_topic.short_upload_schedule_topic.id
    retry_policy   = "RETRY_POLICY_DO_NOT_RETRY"
  }
}

resource "google_service_account" "function_sa" {
  account_id   = "short-upload-fn-sa"
  display_name = "Service Account for Short Upload Cloud Function"
}

resource "google_cloudfunctions2_function_iam_member" "invoker" {
  project        = google_cloudfunctions2_function.short_upload.project
  location       = google_cloudfunctions2_function.short_upload.location
  cloud_function = google_cloudfunctions2_function.short_upload.name
  role           = "roles/cloudfunctions.invoker"
  member         = "serviceAccount:service-${var.project_number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}

resource "google_cloud_run_service_iam_member" "run_invoker" {
  project  = var.project_id
  location = var.region
  service  = google_cloudfunctions2_function.short_upload.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:service-${var.project_number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}

resource "google_cloud_run_service_iam_member" "run_invoker_allusers" {
  project  = var.project_id
  location = var.region
  service  = google_cloudfunctions2_function.short_upload.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

resource "google_project_iam_member" "function_pubsub_subscriber" {
  project = var.project_id
  role    = "roles/pubsub.subscriber"
  member  = "serviceAccount:${google_service_account.function_sa.email}"
}

resource "google_project_iam_member" "function_token_creator" {
  project = var.project_id
  role    = "roles/iam.serviceAccountTokenCreator"
  member  = "serviceAccount:${google_service_account.function_sa.email}"
}
