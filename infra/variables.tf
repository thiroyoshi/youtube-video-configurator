# variables.tf

variable "project_id" {
  description = "GCP Project ID"
  type        = string
  default     = "youtube-video-configurator"
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "asia-northeast1"
}

variable "source_bucket" {
  description = "GCS bucket for function source code"
  type        = string
  default     = "video-converter-src-bucket"
}

variable "project_number" {
  description = "GCP Project Number"
  type        = string
  default     = "589350762095"
}
