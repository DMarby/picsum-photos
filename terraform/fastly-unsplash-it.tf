resource "fastly_service_vcl" "unsplash_redirect" {
  name = "unsplash.it"

  domain {
    name = "unsplash.it"
  }

  backend {
    address        = "127.0.0.1" # Dummy address since we always redirect
    name           = "unsplash-it"
    port           = 80
    use_ssl        = false
  }

  response_object {
    name              = "redirect"
    status            = 301
    response          = "Moved Permanently"
  }

  header {
    name          = "redirect"
    action        = "set"
    type          = "response"
    destination   = "http.Location"
    source        = "\"https://picsum.photos\" req.url"
  }
}

resource "fastly_tls_subscription" "unsplash_redirect" {
  domains               = [for domain in fastly_service_vcl.unsplash_redirect.domain : domain.name]
  certificate_authority = "lets-encrypt"
}

resource "digitalocean_record" "unsplash_tls_domain_validation" {
  depends_on = [fastly_tls_subscription.unsplash_redirect]

  for_each = {
    for challenge in fastly_tls_subscription.unsplash_redirect.managed_dns_challenges :
    trimsuffix(challenge.record_name, "${digitalocean_domain.unsplash.name}.") => challenge
  }

  domain = digitalocean_domain.unsplash.name
  type    = each.value.record_type
  name    = each.key
  value   = "${each.value.record_value}."
  ttl     = 60
}
