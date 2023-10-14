provider "fastly" {
  api_key = var.fastly_api_key
}

data "fastly_tls_configuration" "default" {
  default = true
}
