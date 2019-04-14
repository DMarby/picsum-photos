# domain configuration
resource "digitalocean_domain" "picsum_domain" {
  name = "${var.picsum_domain}"
}
