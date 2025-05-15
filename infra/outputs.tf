# outputs.tf
# 必要なアウトプットがあればここに記述

output "function_name" {
  value = module.convert-starter.function_name
}

output "function_url" {
  value       = module.convert-starter.function_url
  description = "The HTTPS endpoint of the deployed Cloud Function."
}

output "service_account_email" {
  value = module.convert-starter.service_account_email
}
