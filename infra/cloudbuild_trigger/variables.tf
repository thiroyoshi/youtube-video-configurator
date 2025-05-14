variable "trigger_name" {
  description = "Cloud Build Trigger Name"
  type        = string
}

variable "function_name" {
  description = "Cloud Function Name to pass as substitution"
  type        = string
}

variable "cloudbuild_sa_email" {
  description = "Cloud Build Trigger Service Account Email"
  type        = string
}
