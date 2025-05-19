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

variable "x_api_key" {
  description = "X API Key"
  type        = string
  sensitive   = true
}

variable "x_api_secret_key" {
  description = "X API Secret Key"
  type        = string
  sensitive   = true
}

variable "x_access_token" {
  description = "X Access Token"
  type        = string
  sensitive   = true
}

variable "x_access_token_secret" {
  description = "X Access Token Secret"
  type        = string
  sensitive   = true
}

variable "x_user_id" {
  description = "X User ID"
  type        = string
}

variable "youtube_client_id" {
  description = "YouTube Client ID"
  type        = string
  sensitive   = true
}

variable "youtube_client_secret" {
  description = "YouTube Client Secret"
  type        = string
  sensitive   = true
}

variable "youtube_refresh_token" {
  description = "YouTube Refresh Token"
  type        = string
  sensitive   = true
}

variable "playlist_short" {
  description = "YouTube Playlist ID for short videos"
  type        = string
  default     = "PLTSYDCu3sM9LEQ27HYpSlCMrxHyquc-_O"
}

variable "fortnite_season" {
  description = "Current Fortnite season"
  type        = string
  default     = "C6S3"
}
