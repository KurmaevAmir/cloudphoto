output "site_url" {
  value = "https://${var.bucket_name}/"
}


output "bucket" {
  value = yandex_storage_bucket.photos.bucket
}

output "access_key" {
  value = yandex_iam_service_account_static_access_key.cloudphoto.access_key
}

output "secret_key" {
  value     = yandex_iam_service_account_static_access_key.cloudphoto.secret_key
  sensitive = true
}
