resource "digitalocean_kubernetes_cluster" "picsum_k8s" {
  name    = "picsum-k8s"
  region  = var.picsum_digitalocean_region
  version = "1.17.5-do.0"

  node_pool {
    name       = "picsum-pool"
    size       = "s-4vcpu-8gb"
    node_count = 4
  }
}

