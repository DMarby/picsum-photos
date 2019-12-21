resource "digitalocean_database_cluster" "picsum-db" {
  name       = "picsum-db"
  engine     = "pg"
  version    = "11"
  size       = "db-s-1vcpu-1gb"
  region     = var.picsum_digitalocean_region
  node_count = 1
}

