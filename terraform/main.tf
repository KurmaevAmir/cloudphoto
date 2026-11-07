resource "yandex_iam_service_account" "cloudphoto" {
  name        = "cloudphoto-sa-${var.folder_id}"
  description = "Сервисный аккаунт для cloudphoto"
  folder_id   = var.folder_id
}

resource "yandex_iam_service_account_static_access_key" "cloudphoto" {
  service_account_id = yandex_iam_service_account.cloudphoto.id
  description        = "Ключ для программы cloudphoto"
}

resource "yandex_storage_bucket" "photos" {
  bucket    = var.bucket_name
  folder_id = var.folder_id

  https {
    certificate_id = var.cert_id
  }

  website {
    index_document = "index.html"
    error_document = "error.html"
  }

  lifecycle {
    ignore_changes = [acl, grant, anonymous_access_flags]
    precondition {
      condition     = endswith(var.bucket_name, ".${var.domain}")
      error_message = "bucket_name должен быть поддоменом domain"
    }
  }
}

resource "yandex_storage_bucket_iam_binding" "cloudphoto_admin" {
  bucket  = yandex_storage_bucket.photos.bucket
  role    = "storage.admin"
  members = ["serviceAccount:${yandex_iam_service_account.cloudphoto.id}"]
}

data "yandex_dns_zone" "site" {
  dns_zone_id = var.dns_zone_id
}

resource "yandex_dns_recordset" "photos" {
  zone_id = var.dns_zone_id
  name    = "${var.bucket_name}."
  type    = "CNAME"
  ttl     = 600
  data    = ["${var.bucket_name}.website.yandexcloud.net."]

  lifecycle {
    precondition {
      condition     = data.yandex_dns_zone.site.zone == "${var.domain}."
      error_message = "dns_zone_id указывает зону, которая не совпадает с domain."
    }
  }
}
