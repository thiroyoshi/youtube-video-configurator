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
      X_API_KEY             = var.x_api_key
      X_API_SECRET_KEY      = var.x_api_secret_key
      X_ACCESS_TOKEN        = var.x_access_token
      X_ACCESS_TOKEN_SECRET = var.x_access_token_secret
      X_USER_ID             = var.x_user_id
      YOUTUBE_CLIENT_ID     = var.youtube_client_id
      YOUTUBE_CLIENT_SECRET = var.youtube_client_secret
      YOUTUBE_REFRESH_TOKEN = var.youtube_refresh_token
      PLAYLIST_SHORT        = var.playlist_short
      FORTNITE_SEASON       = var.fortnite_season
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
