resource "digitalocean_domain" "unsplash" {
  name = "unsplash.it"
}

resource "digitalocean_record" "caa_issue" {
  domain = digitalocean_domain.unsplash.name

  name  = "@"
  type  = "CAA"
  value = "letsencrypt.org."
  flags = 0
  tag   = "issue"
  ttl   = 60
}

resource "digitalocean_record" "caa_iodef" {
  domain = digitalocean_domain.unsplash.name

  name  = "@"
  type  = "CAA"
  value = "mailto:${var.caa_email}"
  flags = 0
  tag   = "iodef"
  ttl   = 60
}

data "fastly_tls_configuration" "unsplash_redirect" {
  default = true
}

resource "digitalocean_record" "apex" {
  for_each = toset([for each in data.fastly_tls_configuration.unsplash_redirect.dns_records : each.record_value if each.record_type == "A"])

  domain = digitalocean_domain.unsplash.name
  type   = "A"
  name   = "@"
  value  = each.value
  ttl    = 300
}
