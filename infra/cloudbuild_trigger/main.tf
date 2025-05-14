# filepath: infra/cloudbuild_trigger/main.tf

variable "trigger_name" {
  description = "Cloud Build Trigger Name"
  type        = string
}

variable "function_name" {
  description = "Cloud Function Name to pass as substitution"
  type        = string
}

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
}
