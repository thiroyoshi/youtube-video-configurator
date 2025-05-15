provider "google" {
  project = var.project_id
  region  = var.region
}

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 6.35.0"
    }
  }
  required_version = ">= 1.11.0"

  backend "gcs" {
    bucket = "video-converter-state-bucket"
    prefix = "terraform/state"
  }
}
