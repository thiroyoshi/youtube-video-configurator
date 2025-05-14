# variables.tf
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
