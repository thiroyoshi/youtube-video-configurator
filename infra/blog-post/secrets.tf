# Secret Manager secrets for blog-post

# OpenAI API Key secret
resource "google_secret_manager_secret" "openai_api_key" {
  secret_id = "blog-post-openai-api-key"
  replication {
    auto {}
  }
}

# Hatena API Key secret
resource "google_secret_manager_secret" "hatena_api_key" {
  secret_id = "blog-post-hatena-api-key"
  replication {
    auto {}
  }
}

# IAM permissions for Cloud Function service account to access secrets
resource "google_secret_manager_secret_iam_member" "openai_api_key_access" {
  secret_id = google_secret_manager_secret.openai_api_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.function_sa.email}"
}

resource "google_secret_manager_secret_iam_member" "hatena_api_key_access" {
  secret_id = google_secret_manager_secret.hatena_api_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.function_sa.email}"
}