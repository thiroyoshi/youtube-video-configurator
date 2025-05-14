resource "google_cloudbuild_trigger" "this" {
  name        = var.trigger_name
  filename    = "build/cloudbuild.yaml"
  github {
    owner = "thiroyoshi"
    name  = "youtube-video-configurator"
    push {
      branch = "^main$"
    }
  }
  substitutions = {
    _REGION        = "asia-northeast1"
    _FUNCTION_NAME = var.function_name
  }
  service_account = var.cloudbuild_sa_email
}