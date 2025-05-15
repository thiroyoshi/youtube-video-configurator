resource "google_cloudfunctions2_function" "blog_post" {
  name        = "blog-post"
  location    = var.region
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
    min_instance_count = 0
    max_instance_count = 1
    available_memory    = "256M"
    timeout_seconds     = 60
    ingress_settings    = "ALLOW_ALL"
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
