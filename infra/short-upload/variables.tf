variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "project_number" {
  description = "GCP project number"
  type        = string
}

variable "region" {
  description = "GCP region"
  type        = string
}

variable "source_bucket" {
  description = "GCS bucket for source code"
  type        = string
}

variable "short_sha" {
  description = "Short SHA of the commit"
  type        = string
}
