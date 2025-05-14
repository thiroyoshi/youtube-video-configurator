terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 4.0.0"
    }
  }
  required_version = ">= 1.3.0"
}

provider "google" {
  project = var.project_id
  region  = var.region
}

resource "google_service_account" "function_sa" {
  account_id   = "video-converter-fn-sa"
  display_name = "Service Account for Video Converter Cloud Function"
}

resource "google_cloudfunctions2_function" "video_converter" {
  name        = "VideoConverter"
  location    = var.region
  build_config {
    runtime     = "go121"
    entry_point = "VideoConverter"
    source {
      storage_source {
        bucket = var.source_bucket
        object = var.source_object
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
  }
  event_trigger {
    trigger_region = var.region
    event_type     = "google.cloud.pubsub.topic.v1.messagePublished"
    pubsub_topic   = google_pubsub_topic.dummy_topic.id
  }
}

resource "google_pubsub_topic" "dummy_topic" {
  name = "dummy-topic-for-http-trigger"
}

resource "google_cloudfunctions2_function_iam_member" "invoker" {
  project        = google_cloudfunctions2_function.video_converter.project
  location       = google_cloudfunctions2_function.video_converter.location
  cloud_function = google_cloudfunctions2_function.video_converter.name
  role           = "roles/cloudfunctions.invoker"
  member         = "allUsers"
}

variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "asia-northeast1"
}

variable "source_bucket" {
  description = "GCS bucket for function source code"
  type        = string
}

variable "source_object" {
  description = "GCS object (zip) for function source code"
  type        = string
}
