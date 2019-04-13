resource "digitalocean_kubernetes_cluster" "picsum_k8s" {
  name    = "picsum-k8s"
  region  = "ams3"
  version = "1.13.5-do.1"

  node_pool {
    name       = "picsum-pool"
    size       = "s-3vcpu-1gb"
    node_count = 4
  }
}

resource "digitalocean_kubernetes_node_pool" "redis_pool" {
  cluster_id = "${digitalocean_kubernetes_cluster.picsum_k8s.id}"

  name       = "redis-pool"
  size       = "s-2vcpu-4gb"
  node_count = 1
}

resource "digitalocean_certificate" "picsum_k8s_cert" {
  name              = "${var.picsum_domain}"
  type              = "lets_encrypt"
  domains = [
    "${var.picsum_domain}",
    "www.${var.picsum_domain}"
  ]
}
