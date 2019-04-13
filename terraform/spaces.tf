resource "digitalocean_spaces_bucket" "picsum_bucket" {
  name   = "${var.picsum_bucket}"
  region = "${var.picsum_bucket_region}"
  acl = "private"
}
