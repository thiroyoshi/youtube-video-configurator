resource "google_cloudfunctions2_function" "blog_post" {
  name     = "blog-post"
  location = var.region
  build_config {
    runtime     = "go123"
    entry_point = "BlogPost"
    source {
      storage_source {
        bucket = var.source_bucket
        object = "blog-post_${var.short_sha}.zip"
      }
    }
  }
  service_config {
    service_account_email = google_service_account.function_sa.email
    environment_variables = {
      GOOGLE_CLOUD_PROJECT = var.project_id
    }
    secret_environment_variables {
      key        = "OPENAI_API_KEY"
      project_id = var.project_id
      secret     = "blog-post-openai-api-key"
      version    = "latest"
    }
    // HATENA_ID and HATENA_BLOG_ID are now constants in the code
    secret_environment_variables {
      key        = "HATENA_API_KEY"
      project_id = var.project_id
      secret     = "blog-post-hatena-api-key"
      version    = "latest"
    }
    secret_environment_variables {
      key        = "SLACK_WEBHOOK_URL"
      project_id = var.project_id
      secret     = "blog-post-slack-webhook-url"
      version    = "latest"
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
    pubsub_topic   = google_pubsub_topic.blog_post_schedule_topic.id
    retry_policy   = "RETRY_POLICY_DO_NOT_RETRY"
  }
}

resource "google_service_account" "function_sa" {
  account_id   = "blog-post-fn-sa"
  display_name = "Service Account for Blog Post Cloud Function"
}

resource "google_cloudfunctions2_function_iam_member" "invoker" {
  project        = google_cloudfunctions2_function.blog_post.project
  location       = google_cloudfunctions2_function.blog_post.location
  cloud_function = google_cloudfunctions2_function.blog_post.name
  role           = "roles/cloudfunctions.invoker"
  member         = "serviceAccount:service-${var.project_number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}

# Cloud RunのInvoker権限をPub/Subサービスアカウントに付与
resource "google_cloud_run_service_iam_member" "run_invoker" {
  project  = var.project_id
  location = var.region
  service  = google_cloudfunctions2_function.blog_post.name # Cloud Runサービス名はCloud Functions名と同じ
  role     = "roles/run.invoker"
  member   = "serviceAccount:service-${var.project_number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}

# Cloud RunサービスにallUsersのInvoker権限を一時的に付与し、認証エラーの切り分けを行う
resource "google_cloud_run_service_iam_member" "run_invoker_allusers" {
  project  = var.project_id
  location = var.region
  service  = google_cloudfunctions2_function.blog_post.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
