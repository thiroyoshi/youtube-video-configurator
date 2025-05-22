variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "project_number" {
  description = "GCP Project Number"
  type        = string
}

variable "region" {
  description = "GCP region"
  type        = string
}

variable "source_bucket" {
  description = "GCS bucket for function source code"
  type        = string
}

variable "short_sha" {
  description = "Short SHA for artifact versioning"
  type        = string
}
