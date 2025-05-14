variable "project_id" {
  description = "GCP Project ID"
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

variable "convert_starter_service_account_email" {
  description = "convert-starterのサービスアカウントメールアドレス"
  type        = string
}