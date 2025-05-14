# video-converter用Cloud Functions（2nd gen）デプロイ定義
resource "google_cloudfunctions2_function" "video_converter" {
  name        = "video-converter"
  location    = var.region
  build_config {
    runtime     = "go123"
    entry_point = "VideoConverter" # 実際のエントリポイント関数名に合わせて修正
    source {
      storage_source {
        bucket = var.source_bucket
        object = "video-converter.zip"
      }
    }
  }
  service_config {
    service_account_email = google_service_account.video_converter_sa.email
    environment_variables = {
      GOOGLE_CLOUD_PROJECT = var.project_id
    }
    min_instance_count = 0
    max_instance_count = 1
    available_memory    = "256M"
    timeout_seconds     = 60
    ingress_settings    = "ALLOW_ALL"
  }

  labels = {
    "deployment" = "terraform"
  }
}

resource "google_service_account" "video_converter_sa" {
  account_id   = "video-converter-fn-sa"
  display_name = "Service Account for Video Converter Cloud Function"
}

resource "google_cloudfunctions2_function_iam_member" "video_converter_invoker" {
  project        = google_cloudfunctions2_function.video_converter.project
  location       = google_cloudfunctions2_function.video_converter.location
  cloud_function = google_cloudfunctions2_function.video_converter.name
  role           = "roles/cloudfunctions.invoker"
  member         = "allUsers"
}
