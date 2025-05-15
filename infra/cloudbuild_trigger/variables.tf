variable "trigger_name" {
  description = "Cloud Build Trigger Name"
  type        = string
}

variable "function_name" {
  description = "Cloud Function Name to pass as substitution"
  type        = string
}

variable "cloudbuild_sa_id" {
  description = "Cloud Build Trigger Service Account ID"
  type        = string
}

variable "trigger_dir" {
  description = "Cloud Buildトリガーで監視するディレクトリ（例: src/convert-starter）"
  type        = string
}
