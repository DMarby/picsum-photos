# domain configuration
resource "digitalocean_domain" "picsum_domain" {
  name = "${var.picsum_domain}"
}

# CNAME records for subdomains
resource "digitalocean_record" "www" {
  domain = "${digitalocean_domain.picsum_domain.name}"
  type   = "CNAME"
  name   = "www"
  value  = "${digitalocean_domain.picsum_domain.name}."
  ttl    = 60
}

# CAA records
resource "digitalocean_record" "caa_issue" {
  domain = "${digitalocean_domain.picsum_domain.name}"

  name  = "@"
  type  = "CAA"
  value = "letsencrypt.org."
  flags = 128
  tag   = "issue"
  ttl   = 60
}

resource "digitalocean_record" "caa_iodef" {
  domain = "${digitalocean_domain.picsum_domain.name}"

  name  = "@"
  type  = "CAA"
  value = "mailto:${var.caa_email}."
  flags = 128
  tag   = "iodef"
  ttl   = 60
}

# TXT Records
resource "digitalocean_record" "google" {
  domain = "${digitalocean_domain.picsum_domain.name}"
  type   = "TXT"
  name   = "@"
  value  = "google-site-verification=s4mfk0ERh9LpDDZCVex7e1x_w8mUpJMwEgMADaUolQw"
  ttl    = 60
}
