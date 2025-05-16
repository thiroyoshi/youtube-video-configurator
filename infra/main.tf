resource "google_project_service" "serviceusage" {
  project = var.project_id
  service = "serviceusage.googleapis.com"
}

resource "google_project_service" "cloud_scheduler" {
  project    = var.project_id
  service    = "cloudscheduler.googleapis.com"
  depends_on = [google_project_service.serviceusage]
}

resource "google_project_service" "pubsub" {
  project    = var.project_id
  service    = "pubsub.googleapis.com"
  depends_on = [google_project_service.serviceusage]
}

resource "google_service_account" "cloudbuild_sa" {
  account_id   = "cloudbuild-trigger-sa"
  display_name = "Service Account for Cloud Build Trigger"
}

locals {
  cloudbuild_roles = [
    "roles/pubsub.admin",
    "roles/cloudfunctions.admin",
    "roles/cloudscheduler.admin",
    "roles/resourcemanager.projectIamAdmin",
    "roles/storage.admin",           // GCS操作用
    "roles/iam.serviceAccountUser",  // サービスアカウント指定デプロイ用
    "roles/logging.logWriter",       // Cloud Logging書き込み権限
    "roles/cloudbuild.builds.editor" // Cloud Build Trigger管理用
  ]
}

resource "google_project_iam_member" "cloudbuild_sa_roles" {
  for_each = toset(local.cloudbuild_roles)
  project  = var.project_id
  role     = each.value
  member   = "serviceAccount:${google_service_account.cloudbuild_sa.email}"
}

resource "time_sleep" "wait_for_scheduler_api" {
  depends_on      = [google_project_service.cloud_scheduler]
  create_duration = "30s"
}

module "convert-starter_deploy_trigger" {
  source           = "./cloudbuild_trigger"
  trigger_name     = "convert-starter-deploy-trigger"
  function_name    = "convert-starter"
  trigger_dir      = "src/convert-starter"
  cloudbuild_sa_id = google_service_account.cloudbuild_sa.id
}

module "video-converter_deploy_trigger" {
  source           = "./cloudbuild_trigger"
  trigger_name     = "video-converter-deploy-trigger"
  function_name    = "video-converter"
  trigger_dir      = "src/video-converter"
  cloudbuild_sa_id = google_service_account.cloudbuild_sa.id
}

module "blog-post_deploy_trigger" {
  source           = "./cloudbuild_trigger"
  trigger_name     = "blog-post-deploy-trigger"
  function_name    = "blog-post"
  trigger_dir      = "src/blog-post"
  cloudbuild_sa_id = google_service_account.cloudbuild_sa.id
}

module "convert-starter" {
  source         = "./convert-starter"
  project_id     = var.project_id
  project_number = var.project_number
  region         = var.region
  source_bucket  = var.source_bucket
  short_sha      = var.short_sha
  depends_on     = [time_sleep.wait_for_scheduler_api, google_project_service.pubsub]
}

module "video-converter" {
  source                                = "./video-converter"
  project_id                            = var.project_id
  region                                = var.region
  source_bucket                         = var.source_bucket
  convert_starter_service_account_email = module.convert-starter.service_account_email
  short_sha                             = var.short_sha
}

module "blog-post" {
  source         = "./blog-post"
  project_id     = var.project_id
  project_number = var.project_number
  region         = var.region
  source_bucket  = var.source_bucket
  short_sha      = var.short_sha
  depends_on     = [time_sleep.wait_for_scheduler_api, google_project_service.pubsub]
}

module "blog_post_secrets" {
  source     = "./secret-manager"
  project_id = var.project_id
  secret_ids = [
    "blog-post-openai-api-key",
    "blog-post-hatena-id",
    "blog-post-hatena-blog-id",
    "blog-post-hatena-api-key",
    "blog-post-slack-webhook-url"
  ]
  service_account_email = module.blog-post.service_account_email
  depends_on            = [module.blog-post]
}
