resource "digitalocean_kubernetes_cluster" "picsum_k8s" {
  name    = "picsum-k8s"
  region  = "ams3"
  version = "1.13.5-do.1"

  node_pool {
    name       = "picsum-pool"
    size       = "s-4vcpu-8gb"
    node_count = 2
  }
}
