variable "folder_id" {
  description = "ID каталога Yandex Cloud"
  type        = string
}

variable "bucket_name" {
  description = "Имя бакета"
  type        = string
}

variable "domain" {
  description = "Домен (DNS-зона), в которой добавляются записи сайта"
  type        = string
}

variable "cert_id" {
  description = "ID сертификата"
  type        = string
}

variable "dns_zone_id" {
  description = "ID DNS-зоны"
  type        = string
}
