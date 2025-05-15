resource "google_cloudfunctions2_function" "convert_starter" {
  name     = "convert-starter"
  location = var.region
  build_config {
    runtime     = "go123"
    entry_point = "convertStarter"
    source {
      storage_source {
        bucket = var.source_bucket
        object = "convert-starter_${var.short_sha}.zip"
      }
    }
  }
  service_config {
    service_account_email = google_service_account.function_sa.email
    environment_variables = {
      GOOGLE_CLOUD_PROJECT = var.project_id
    }
    min_instance_count             = 0
    max_instance_count             = 1
    available_memory               = "256M"
    timeout_seconds                = 60
    ingress_settings               = "ALLOW_ALL"
    all_traffic_on_latest_revision = true
  }

  event_trigger {
    trigger_region = var.region
    event_type     = "google.cloud.pubsub.topic.v1.messagePublished"
    pubsub_topic   = google_pubsub_topic.convert_starter_schedule_topic.id
    retry_policy   = "RETRY_POLICY_DO_NOT_RETRY"
  }
}

resource "google_service_account" "function_sa" {
  account_id   = "convert-starter-fn-sa"
  display_name = "Service Account for Convert Starter Cloud Function"
}

# Cloud FunctionsのInvoker権限をPub/Subサービスアカウントに付与
resource "google_cloudfunctions2_function_iam_member" "invoker" {
  project        = google_cloudfunctions2_function.convert_starter.project
  location       = google_cloudfunctions2_function.convert_starter.location
  cloud_function = google_cloudfunctions2_function.convert_starter.name
  role           = "roles/cloudfunctions.invoker"
  member         = "serviceAccount:service-${var.project_number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}

# Cloud RunのInvoker権限をPub/Subサービスアカウントに付与
resource "google_cloud_run_service_iam_member" "run_invoker" {
  project  = var.project_id
  location = var.region
  service  = google_cloudfunctions2_function.convert_starter.name # Cloud Runサービス名はCloud Functions名と同じ
  role     = "roles/run.invoker"
  member   = "serviceAccount:service-${var.project_number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}

# Cloud RunサービスにallUsersのInvoker権限を一時的に付与し、認証エラーの切り分けを行う
resource "google_cloud_run_service_iam_member" "run_invoker_allusers" {
  project  = var.project_id
  location = var.region
  service  = google_cloudfunctions2_function.convert_starter.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# Cloud FunctionsのサービスアカウントにPub/Subの権限を付与
resource "google_project_iam_member" "function_pubsub_subscriber" {
  project = var.project_id
  role    = "roles/pubsub.subscriber"
  member  = "serviceAccount:${google_service_account.function_sa.email}"
}

# Cloud FunctionsのサービスアカウントにToken Creatorの権限を付与
resource "google_project_iam_member" "function_token_creator" {
  project = var.project_id
  role    = "roles/iam.serviceAccountTokenCreator"
  member  = "serviceAccount:${google_service_account.function_sa.email}"
}
