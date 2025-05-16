resource "google_project_service" "secretmanager" {
  project = var.project_id
  service = "secretmanager.googleapis.com"
}

resource "google_secret_manager_secret" "secrets" {
  for_each = toset(var.secret_ids)

  project   = var.project_id
  secret_id = each.value

  replication {
    auto {}  # Use Google-managed keys for replication
  }

  depends_on = [google_project_service.secretmanager]
}

resource "google_secret_manager_secret_version" "secret_versions" {
  for_each = google_secret_manager_secret.secrets

  secret      = each.value.id
  secret_data = ""  # Empty initial value

  lifecycle {
    ignore_changes = [
      secret_data, # Ignore changes to secret_data since it will be managed outside Terraform
    ]
  }
}

resource "google_secret_manager_secret_iam_member" "secret_access" {
  for_each = google_secret_manager_secret.secrets

  project   = var.project_id
  secret_id = each.value.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${var.service_account_email}"
}
