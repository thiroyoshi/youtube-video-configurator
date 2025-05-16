variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "secret_ids" {
  description = "List of secret IDs to create"
  type        = list(string)
}

variable "service_account_email" {
  description = "Service account email that needs access to the secrets"
  type        = string
}
