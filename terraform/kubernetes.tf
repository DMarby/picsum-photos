resource "digitalocean_kubernetes_cluster" "picsum_k8s" {
  name    = "picsum-k8s"
  region  = var.picsum_digitalocean_region
  version = "1.24.4-do.0"

  auto_upgrade = true

  node_pool {
    name       = "picsum-pool"
    size       = "s-4vcpu-8gb"
    node_count = 4
  }

  lifecycle {
    ignore_changes = [
      version
    ]
  }
}

resource "digitalocean_container_registry" "picsum_registry" {
  name                   = "picsum-registry"
  subscription_tier_slug = "basic"
  region                 = var.picsum_digitalocean_region
}

provider "kubernetes" {
  host  = digitalocean_kubernetes_cluster.picsum_k8s.endpoint
  token = digitalocean_kubernetes_cluster.picsum_k8s.kube_config[0].token
  cluster_ca_certificate = base64decode(
    digitalocean_kubernetes_cluster.picsum_k8s.kube_config[0].cluster_ca_certificate
  )
}

resource "digitalocean_container_registry_docker_credentials" "picsum_registry" {
  registry_name = "picsum-registry"
}
resource "kubernetes_secret" "picsum_registry" {
  metadata {
    name = "picsum-registry"
  }

  data = {
    ".dockerconfigjson" = digitalocean_container_registry_docker_credentials.picsum_registry.docker_credentials
  }

  type = "kubernetes.io/dockerconfigjson"
}
